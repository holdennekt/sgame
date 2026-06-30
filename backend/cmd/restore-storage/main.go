package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	gcsstorage "cloud.google.com/go/storage"
	"github.com/holdennekt/sgame/backend/pkg/envvar"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "list soft-deleted objects without restoring them")
	since := flag.Duration("since", 24*time.Hour, "restore objects soft-deleted within this duration (e.g. 2h, 24h, 48h)")
	flag.Parse()

	_ = godotenv.Load()
	ctx := context.Background()

	bucketName := envvar.GetEnvVar("BUCKET_NAME")

	gcsClient, err := gcsstorage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = gcsClient.Close() }()

	bucket := gcsClient.Bucket(bucketName)
	cutoff := time.Now().Add(-*since)

	log.Printf("Listing soft-deleted objects in bucket %q deleted after %s...", bucketName, cutoff.Format(time.RFC3339))

	it := bucket.Objects(ctx, &gcsstorage.Query{SoftDeleted: true})

	var found, restored, skipped int
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("listing soft-deleted objects: %v", err)
		}

		found++

		if attrs.SoftDeleteTime.Before(cutoff) {
			skipped++
			continue
		}

		if *dryRun {
			fmt.Printf("[dry-run] restore %s (generation %d, soft-deleted %s)\n",
				attrs.Name, attrs.Generation, attrs.SoftDeleteTime.Format(time.RFC3339))
			restored++
			continue
		}

		_, err = bucket.Object(attrs.Name).
			Generation(attrs.Generation).
			SoftDeleted().
			Restore(ctx, &gcsstorage.RestoreOptions{CopySourceACL: true})
		if err != nil {
			log.Printf("ERROR restoring %s (generation %d): %v", attrs.Name, attrs.Generation, err)
			continue
		}

		log.Printf("Restored %s", attrs.Name)
		restored++
	}

	log.Println("=== Summary ===")
	log.Printf("Soft-deleted objects found: %d", found)
	log.Printf("Too old (before cutoff):    %d", skipped)
	if *dryRun {
		log.Printf("Would restore:              %d", restored)
	} else {
		log.Printf("Restored:                   %d", restored)
	}
}
