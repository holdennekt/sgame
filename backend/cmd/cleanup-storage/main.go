package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	gcsstorage "cloud.google.com/go/storage"
	"github.com/holdennekt/sgame/backend/pkg/envvar"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/iterator"
)

type attachmentDoc struct {
	Key string `bson:"key"`
}

type commentDoc struct {
	Attachment *attachmentDoc `bson:"attachment"`
}

type questionDoc struct {
	Attachment *attachmentDoc `bson:"attachment"`
	Comment    *commentDoc    `bson:"comment"`
}

type categoryDoc struct {
	Questions []questionDoc `bson:"questions"`
}

type roundDoc struct {
	Categories []categoryDoc `bson:"categories"`
}

type finalRoundQuestionDoc struct {
	Attachment *attachmentDoc `bson:"attachment"`
	Comment    *commentDoc    `bson:"comment"`
}

type finalRoundCategoryDoc struct {
	Question finalRoundQuestionDoc `bson:"question"`
}

type finalRoundDoc struct {
	Categories []finalRoundCategoryDoc `bson:"categories"`
}

type packDoc struct {
	Rounds     []roundDoc    `bson:"rounds"`
	FinalRound finalRoundDoc `bson:"finalRound"`
}

type userDoc struct {
	Avatar *string `bson:"avatar"`
}

var packAttachmentProjection = bson.M{
	"rounds.categories.questions.attachment.key":            1,
	"rounds.categories.questions.comment.attachment.key":    1,
	"finalRound.categories.question.attachment.key":         1,
	"finalRound.categories.question.comment.attachment.key": 1,
}

func collectKeysFromPackDocs(ctx context.Context, cur *mongo.Cursor) (map[string]struct{}, error) {
	defer cur.Close(ctx)
	keys := make(map[string]struct{})
	for cur.Next(ctx) {
		var pack packDoc
		if err := cur.Decode(&pack); err != nil {
			return nil, err
		}
		for _, round := range pack.Rounds {
			for _, cat := range round.Categories {
				for _, q := range cat.Questions {
					if q.Attachment != nil && q.Attachment.Key != "" {
						keys[q.Attachment.Key] = struct{}{}
					}
					if q.Comment != nil && q.Comment.Attachment != nil && q.Comment.Attachment.Key != "" {
						keys[q.Comment.Attachment.Key] = struct{}{}
					}
				}
			}
		}
		for _, cat := range pack.FinalRound.Categories {
			if cat.Question.Attachment != nil && cat.Question.Attachment.Key != "" {
				keys[cat.Question.Attachment.Key] = struct{}{}
			}
			if cat.Question.Comment != nil && cat.Question.Comment.Attachment != nil && cat.Question.Comment.Attachment.Key != "" {
				keys[cat.Question.Comment.Attachment.Key] = struct{}{}
			}
		}
	}
	return keys, cur.Err()
}

func collectPackKeys(ctx context.Context, db *mongo.Database) (map[string]struct{}, error) {
	cur, err := db.Collection("packs").Find(ctx, bson.M{}, options.Find().SetProjection(packAttachmentProjection))
	if err != nil {
		return nil, err
	}
	return collectKeysFromPackDocs(ctx, cur)
}

func collectPackDraftKeys(ctx context.Context, db *mongo.Database) (map[string]struct{}, error) {
	cur, err := db.Collection("pack_drafts").Find(ctx, bson.M{}, options.Find().SetProjection(packAttachmentProjection))
	if err != nil {
		return nil, err
	}
	return collectKeysFromPackDocs(ctx, cur)
}

func collectAvatarKeys(ctx context.Context, db *mongo.Database) (map[string]struct{}, error) {
	cur, err := db.Collection("users").Find(
		ctx,
		bson.M{"avatar": bson.M{"$ne": nil}},
		options.Find().SetProjection(bson.M{"avatar": 1}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	keys := make(map[string]struct{})
	for cur.Next(ctx) {
		var user userDoc
		if err := cur.Decode(&user); err != nil {
			return nil, err
		}
		if user.Avatar != nil && *user.Avatar != "" {
			keys[*user.Avatar] = struct{}{}
		}
	}
	return keys, cur.Err()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	dryRun := flag.Bool("dry-run", false, "list orphaned objects without deleting them")
	flag.Parse()

	godotenv.Load()
	ctx := context.Background()

	mongoURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s?authSource=admin",
		envvar.GetEnvVar("MONGO_ROOT_USER"),
		envvar.GetEnvVar("MONGO_ROOT_PASSWORD"),
		envvar.GetEnvVar("MONGO_HOST"),
		envvar.GetEnvVar("MONGO_PORT"),
		envvar.GetEnvVar("MONGO_NAME"),
	)
	mongoConn, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoConn.Disconnect(ctx)
	db := mongoConn.Database(envvar.GetEnvVar("MONGO_NAME"))

	log.Println("Collecting referenced keys from MongoDB...")

	packKeys, err := collectPackKeys(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Pack attachment keys: %d", len(packKeys))

	packDraftKeys, err := collectPackDraftKeys(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Pack draft attachment keys: %d", len(packDraftKeys))

	avatarKeys, err := collectAvatarKeys(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Avatar keys: %d", len(avatarKeys))

	referenced := make(map[string]struct{}, len(packKeys)+len(packDraftKeys)+len(avatarKeys))
	for k := range packKeys {
		referenced[k] = struct{}{}
	}
	for k := range packDraftKeys {
		referenced[k] = struct{}{}
	}
	for k := range avatarKeys {
		referenced[k] = struct{}{}
	}
	log.Printf("Total referenced keys: %d", len(referenced))

	bucketName := envvar.GetEnvVar("BUCKET_NAME")
	provider := os.Getenv("STORAGE_PROVIDER")

	var totalObjects, orphaned int
	var totalSize, orphanedSize int64

	if provider == "gcs" {
		gcsClient, err := gcsstorage.NewClient(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer gcsClient.Close()

		log.Printf("Listing objects in GCS bucket %q...", bucketName)

		it := gcsClient.Bucket(bucketName).Objects(ctx, nil)
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			totalObjects++
			totalSize += attrs.Size

			if _, ok := referenced[attrs.Name]; !ok {
				orphaned++
				orphanedSize += attrs.Size
				if *dryRun {
					fmt.Printf("[dry-run] %s (%s)\n", attrs.Name, formatBytes(attrs.Size))
				} else {
					log.Printf("Deleting %s...", attrs.Name)
					if err := gcsClient.Bucket(bucketName).Object(attrs.Name).Delete(ctx); err != nil {
						log.Printf("ERROR deleting %s: %v", attrs.Name, err)
					}
				}
			}
		}
	} else {
		minioClient, err := minio.New(envvar.GetEnvVar("MINIO_ENDPOINT"), &minio.Options{
			Creds: credentials.NewStaticV4(
				envvar.GetEnvVar("MINIO_ROOT_USER"),
				envvar.GetEnvVar("MINIO_ROOT_PASSWORD"),
				"",
			),
			Secure: envvar.GetEnvVarBool("MINIO_USE_SSL"),
		})
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Listing objects in MinIO bucket %q...", bucketName)

		for obj := range minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true}) {
			if obj.Err != nil {
				log.Fatal(obj.Err)
			}

			totalObjects++
			totalSize += obj.Size

			if _, ok := referenced[obj.Key]; !ok {
				orphaned++
				orphanedSize += obj.Size
				if *dryRun {
					fmt.Printf("[dry-run] %s (%s)\n", obj.Key, formatBytes(obj.Size))
				} else {
					log.Printf("Deleting %s...", obj.Key)
					if err := minioClient.RemoveObject(ctx, bucketName, obj.Key, minio.RemoveObjectOptions{}); err != nil {
						log.Printf("ERROR deleting %s: %v", obj.Key, err)
					}
				}
			}
		}
	}

	action := "deleted"
	if *dryRun {
		action = "would delete"
	}

	log.Println("=== Summary ===")
	log.Printf("Total objects:         %d (%s)", totalObjects, formatBytes(totalSize))
	log.Printf("Orphaned (%s): %d (%s)", action, orphaned, formatBytes(orphanedSize))
	log.Printf("Referenced (kept):     %d (%s)", totalObjects-orphaned, formatBytes(totalSize-orphanedSize))
}
