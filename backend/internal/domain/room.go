package domain

import (
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/holdennekt/sgame/pkg/custerr"
)

const LOBBY = "lobby"
const ROOM_PREFIX = "room:"
const LOCK_PREFIX = "lock:room:"
const INTERNAL_POSTFIX = ":internal"

type Room struct {
	RoomDTO               `bson:"inline"`
	Id                    string                `json:"id" bson:"_id"`
	PackPreview           PackPreview           `json:"packPreview" bson:"packPreview"`
	CreatedBy             string                `json:"createdBy" bson:"createdBy"`
	Host                  *Host                 `json:"host" bson:"host"`
	Players               []Player              `json:"players" bson:"players"`
	BanList               []string              `json:"banList" bson:"banList"`
	State                 RoomState             `json:"state" bson:"state"`
	CurrentRoundName      *string               `json:"currentRoundName" bson:"currentRoundName"`
	CurrentRoundQuestions CurrentRoundQuestions `json:"currentRoundQuestions" bson:"currentRoundQuestions"`
	CurrentPlayer         *string               `json:"currentPlayer" bson:"currentPlayer"`
	CurrentQuestion       *CurrentQuestion      `json:"currentQuestion" bson:"currentQuestion"`
	AnsweringPlayer       *AnsweringPlayer      `json:"answeringPlayer" bson:"answeringPlayer"`
	AllowedToAnswer       []string              `json:"allowedToAnswer" bson:"allowedToAnswer"`
	FinalRoundState       *FinalRoundState      `json:"finalRoundState" bson:"finalRoundState"`
	PausedState           PausedState           `json:"pausedState" bson:"pausedState"`
}

type RoomDTO struct {
	Name    string      `json:"name" bson:"name" binding:"min=1,max=50"`
	PackId  string      `json:"packId" bson:"packId" binding:"required"`
	Options roomOptions `json:"options" bson:"options"`
}

type roomOptions struct {
	MaxPlayers                int         `json:"maxPlayers" bson:"maxPlayers" binding:"min=1,max=10"`
	Type                      PrivacyType `json:"type" bson:"type" binding:"oneof=public private"`
	Password                  *string     `json:"password" bson:"password" binding:"omitnil,min=4,max=16"`
	QuestionThinkingTime      int         `json:"questionThinkingTime" bson:"questionThinkingTime" binding:"min=1,max=30"`
	AnswerThinkingTime        int         `json:"answerThinkingTime" bson:"answerThinkingTime" binding:"min=1,max=30"`
	QuestionThinkingTimeFinal int         `json:"questionThinkingTimeFinal" bson:"questionThinkingTimeFinal" binding:"min=1,max=120"`
	FalseStartAllowed         bool        `json:"falseStartAllowed" bson:"falseStartAllowed"`
}

type PrivacyType string

const (
	Public  PrivacyType = "public"
	Private PrivacyType = "private"
)

type RoomState string

const (
	WaitingForStart             RoomState = "waiting_for_start"
	SelectingQuestion           RoomState = "selecting_question"
	RevealingQuestion           RoomState = "revealing_question"
	ShowingQuestion             RoomState = "showing_question"
	Answering                   RoomState = "answering"
	Betting                     RoomState = "betting"
	Passing                     RoomState = "passing"
	SelectingFinalRoundCategory RoomState = "selecting_final_round_category"
	FinalRoundBetting           RoomState = "final_round_betting"
	ShowingFinalRoundQuestion   RoomState = "showing_final_round_question"
	ValidatingFinalRoundAnswers RoomState = "validating_final_round_answers"
	GameOver                    RoomState = "game_over"
)

type CurrentRoundQuestions map[string][]BoardQuestion
type BoardQuestion struct {
	Index         int  `json:"index" bson:"index"`
	Value         int  `json:"value" bson:"value"`
	HasBeenPlayed bool `json:"hasBeenPlayed" bson:"hasBeenPlayed"`
}

type CurrentQuestion struct {
	Question
	TimerLastProgress float64   `json:"timerLastProgress" bson:"timerLastProgress"`
	TimerStartsAt     time.Time `json:"timerStartsAt" bson:"timerStartsAt"`
	TimerEndsAt       time.Time `json:"timerEndsAt" bson:"timerEndsAt"`
}

type AnsweringPlayer struct {
	Id            string    `json:"id" bson:"id"`
	TimerStartsAt time.Time `json:"timerStartsAt" bson:"timerStartsAt"`
	TimerEndsAt   time.Time `json:"timerEndsAt" bson:"timerEndsAt"`
}

type FinalRoundState struct {
	AvailableCategories map[string]bool     `json:"availableCategories" bson:"availableCategories"`
	Question            *FinalRoundQuestion `json:"question" bson:"question"`
	Players             []string            `json:"players" bson:"players"`
	PlayersAnswers      map[string]string   `json:"playersAnswers" bson:"playersAnswers"`
	TimerEndsAt         *time.Time          `json:"timerEndsAt" bson:"timerEndsAt"`
}

type PausedState struct {
	Paused   bool       `json:"paused" bson:"paused"`
	PausedAt *time.Time `json:"pausedAt" bson:"pausedAt"`
}

func (r *Room) IsUserHost(userId string) bool {
	if r.Host == nil {
		return false
	}
	return userId == r.Host.Id
}

func (r *Room) IsUserPlayer(userId string) bool {
	return slices.ContainsFunc(r.Players, func(player Player) bool {
		return userId == player.Id
	})
}

func (r *Room) IsUserIn(userId string) bool {
	return r.IsUserHost(userId) || r.IsUserPlayer(userId)
}

func (r *Room) IsUserBanned(userId string) bool {
	return slices.ContainsFunc(r.BanList, func(id string) bool {
		return userId == id
	})
}

func (r *Room) GetProjection(userId string) any {
	if r.IsUserIn(userId) {
		if r.IsUserHost(userId) {
			return NewHostRoom(r)
		}
		return NewPlayerRoom(r)
	}
	return NewRoomLobby(r)
}

func (r *Room) StartGame(pack *Pack) {
	r.StartNextRegularRound(pack)
	r.CurrentPlayer = &r.Players[rand.Intn(len(r.Players))].Id
}

func (r *Room) StartNextRegularRound(pack *Pack) bool {
	var nextRoundIndex int
	if r.State == WaitingForStart {
		nextRoundIndex = 0
	} else {
		currentRoundIndex := slices.IndexFunc(pack.Rounds, func(round Round) bool {
			return *r.CurrentRoundName == round.Name
		})
		nextRoundIndex = currentRoundIndex + 1
	}
	if nextRoundIndex < len(pack.Rounds) {
		nextRound := pack.Rounds[nextRoundIndex]
		r.CurrentRoundName = &nextRound.Name
		r.CurrentRoundQuestions = nextRound.getCurrentRoundQuestions()
		r.State = SelectingQuestion
		return true
	}
	return false
}

func (r *Room) SelectQuestion(userId string, pack *Pack, category string, index int) error {
	if r.State != SelectingQuestion {
		return custerr.NewConflictErr("can not select question now")
	}
	if userId != *r.CurrentPlayer && userId != r.Host.Id {
		return custerr.NewForbiddenErr("not allowed to select question")
	}
	if r.CurrentRoundQuestions[category][index].HasBeenPlayed {
		return custerr.NewConflictErr("question has already been played")
	}

	question, err := pack.getQuestion(*r.CurrentRoundName, category, index)
	if err != nil {
		return err
	}

	r.CurrentQuestion = &CurrentQuestion{Question: question}
	r.CurrentRoundQuestions[category][index].HasBeenPlayed = true

	switch question.Type {
	case Regular:
		allowedToAnswer := make([]string, len(r.Players))
		for i, player := range r.Players {
			allowedToAnswer[i] = player.Id
		}
		r.revealRegularQuestion(allowedToAnswer)
	case CatInBag:
		canPassTo := make([]Player, 0)
		for _, p := range r.Players {
			if p.Id != userId && p.IsConnected {
				canPassTo = append(canPassTo, p)
			}
		}
		if len(canPassTo) == 0 {
			r.startNonRegularQuestion(userId)
		} else if len(canPassTo) == 1 {
			r.startNonRegularQuestion(canPassTo[0].Id)
		} else {
			r.State = Passing
		}
	case Auction:
		canBet := make([]Player, 0)
		for _, p := range r.Players {
			if p.Score > 0 {
				canBet = append(canBet, p)
			}
		}
		if len(canBet) > 0 {
			r.State = Betting
		} else {
			allowedToAnswer := make([]string, len(r.Players))
			for i, player := range r.Players {
				allowedToAnswer[i] = player.Id
			}
			r.revealRegularQuestion(allowedToAnswer)
		}
	}

	return nil
}

func (r *Room) revealRegularQuestion(allowedToAnswer []string) {
	r.AllowedToAnswer = allowedToAnswer
	revealingDuration := r.CurrentQuestion.Question.GetRevealingDuration()
	r.CurrentQuestion.TimerStartsAt = time.Now().Add(revealingDuration)
	r.State = RevealingQuestion
}

func (r *Room) StartRegularQuestion() {
	thinkingDuration := time.Duration(r.Options.QuestionThinkingTime) * time.Second
	r.CurrentQuestion.TimerEndsAt = r.CurrentQuestion.TimerStartsAt.Add(thinkingDuration)
	r.CurrentQuestion.TimerLastProgress = 1
	r.State = ShowingQuestion
}

func (r *Room) startNonRegularQuestion(allowedToAnswer string) {
	r.CurrentPlayer = &allowedToAnswer
	now := time.Now()
	revealingDuration := r.CurrentQuestion.Question.GetRevealingDuration()
	thinkingDuration := time.Duration(r.Options.AnswerThinkingTime) * time.Second
	r.AnsweringPlayer = &AnsweringPlayer{
		Id:            allowedToAnswer,
		TimerStartsAt: now,
		TimerEndsAt:   now.Add(revealingDuration + thinkingDuration),
	}
	r.State = Answering
}

func (r *Room) SubmitAnswer(userId string) error {
	if r.State != RevealingQuestion && r.State != ShowingQuestion {
		return custerr.NewConflictErr("can not submit answer now")
	}
	if !slices.Contains(r.AllowedToAnswer, userId) {
		return custerr.NewConflictErr("not allowed to submit answer")
	}
	if r.State == RevealingQuestion && !r.Options.FalseStartAllowed {
		return custerr.NewConflictErr("can not submit answer now")
	}

	now := time.Now()
	thinkingDuration := time.Duration(r.Options.AnswerThinkingTime) * time.Second
	r.AnsweringPlayer = &AnsweringPlayer{
		Id:            userId,
		TimerStartsAt: now,
		TimerEndsAt:   now.Add(thinkingDuration),
	}
	r.CurrentPlayer = &userId
	r.AllowedToAnswer = slices.DeleteFunc(r.AllowedToAnswer, func(playerId string) bool {
		return userId == playerId
	})
	r.State = Answering
	return nil
}

func (r *Room) ValidateAnswer(userId string, isCorrect bool) error {
	if r.State != Answering {
		return custerr.NewConflictErr("can not validate answer now")
	}
	if !r.IsUserHost(userId) && userId != SYSTEM {
		return custerr.NewForbiddenErr("not allowed to validate answer")
	}

	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return r.AnsweringPlayer.Id == p.Id
	})
	questionValue := r.CurrentQuestion.Value
	if betAmount := r.Players[playerIndex].BetAmount; betAmount != nil {
		questionValue = *betAmount
	}
	if isCorrect {
		r.Players[playerIndex].Score += questionValue
	} else {
		r.Players[playerIndex].Score -= questionValue
	}

	if isCorrect || len(r.AllowedToAnswer) == 0 {
		r.EndQuestion()
	} else {
		r.continueRegularQuestion()
	}
	return nil
}

func (r *Room) continueRegularQuestion() {
	now := time.Now()
	answerDuration := now.Sub(r.AnsweringPlayer.TimerStartsAt)
	if r.AnsweringPlayer.TimerStartsAt.Before(r.CurrentQuestion.TimerStartsAt) {
		r.CurrentQuestion.TimerStartsAt = r.CurrentQuestion.TimerStartsAt.Add(answerDuration)
		r.State = RevealingQuestion
	} else {
		r.CurrentQuestion.TimerEndsAt = r.CurrentQuestion.TimerEndsAt.Add(answerDuration)
		questionDurationRemained := r.CurrentQuestion.TimerEndsAt.Sub(now)
		r.CurrentQuestion.TimerLastProgress = float64(questionDurationRemained) /
			float64((time.Duration(r.Options.QuestionThinkingTime) * time.Second))
		r.State = ShowingQuestion
	}
	r.AnsweringPlayer = nil
}

func (r *Room) PassQuestion(fromUserId string, toUserId string) error {
	if r.State != Passing {
		return custerr.NewConflictErr("can not pass question now")
	}
	toUserIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return p.Id == toUserId
	})
	if toUserIndex == -1 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no player with id \"%s\" in room", toUserId))
	}
	if *r.CurrentPlayer != fromUserId && r.Host.Id != fromUserId {
		return custerr.NewConflictErr("not allowed to pass question")
	}
	if *r.CurrentPlayer == toUserId {
		return custerr.NewConflictErr("can not pass question to current player")
	}
	if !r.Players[toUserIndex].IsConnected {
		return custerr.NewConflictErr("can not pass question to disconnected player")
	}

	r.startNonRegularQuestion(toUserId)
	return nil
}

func (r *Room) PassQuestionAuto() {
	var passTo string
	canPassTo := make([]Player, 0)
	for _, p := range r.Players {
		if p.Id != *r.CurrentPlayer && p.IsConnected {
			canPassTo = append(canPassTo, p)
		}
	}
	if len(canPassTo) == 0 {
		passTo = *r.CurrentPlayer
	} else {
		passTo = canPassTo[rand.Intn(len(canPassTo))].Id
	}
	r.startNonRegularQuestion(passTo)
}

func (r *Room) PlaceBet(userId string, amount int) error {
	if r.State != Betting {
		return custerr.NewConflictErr("can not place bet now")
	}
	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return userId == p.Id
	})
	alreadyBet := r.Players[playerIndex].BetAmount != nil
	if alreadyBet {
		return custerr.NewConflictErr("can not place bet again")
	}
	// betIncreased := r.Players[playerIndex].BetAmount != nil && *r.Players[playerIndex].BetAmount < amount
	insufficientScore := amount > r.Players[playerIndex].Score || amount < 0
	if insufficientScore {
		return custerr.NewConflictErr("insufficient bet size")
	}

	r.Players[playerIndex].BetAmount = &amount

	canBet := make([]Player, 0)
	for _, p := range r.Players {
		if p.Score > 0 {
			canBet = append(canBet, p)
		}
	}
	allBet := !slices.ContainsFunc(canBet, func(p Player) bool {
		return p.BetAmount == nil
	})
	if allBet {
		playerMaxBet := canBet[0]
		for _, p := range canBet[1:] {
			if *playerMaxBet.BetAmount < *p.BetAmount {
				playerMaxBet = p
			}
		}
		if *playerMaxBet.BetAmount > 0 {
			r.startNonRegularQuestion(playerMaxBet.Id)
		} else {
			r.EndQuestion()
		}
	}
	return nil
}

func (r *Room) PlaceBetsAuto() {
	canBet := make([]Player, 0)
	for _, p := range r.Players {
		if p.Score > 0 {
			canBet = append(canBet, p)
		}
	}
	for i, player := range canBet {
		if player.BetAmount == nil {
			zero := 0
			canBet[i].BetAmount = &zero
		}
	}
	playerMaxBet := canBet[0]
	for _, p := range canBet[1:] {
		if *playerMaxBet.BetAmount < *p.BetAmount {
			playerMaxBet = p
		}
	}
	if *playerMaxBet.BetAmount > 0 {
		r.startNonRegularQuestion(playerMaxBet.Id)
	} else {
		r.EndQuestion()
	}
}

func (r *Room) EndQuestion() {
	r.CurrentQuestion = nil
	r.AnsweringPlayer = nil
	r.AllowedToAnswer = make([]string, 0)
	for i := range r.Players {
		r.Players[i].BetAmount = nil
	}
	r.State = SelectingQuestion
}

func (r *Room) AnyAvailableQuestions() bool {
	for _, questions := range r.CurrentRoundQuestions {
		notPlayedIndex := slices.IndexFunc(questions, func(bq BoardQuestion) bool {
			return !bq.HasBeenPlayed
		})
		if notPlayedIndex != -1 {
			return true
		}
	}
	return false
}

func (r *Room) StartFinalRound(pack *Pack) bool {
	r.CurrentRoundName = nil
	r.CurrentRoundQuestions = nil
	r.CurrentPlayer = nil
	r.CurrentQuestion = nil
	r.AnsweringPlayer = nil
	r.AllowedToAnswer = nil

	finalRoundPlayers := make([]string, 0)
	for _, player := range r.Players {
		if player.Score > 0 {
			finalRoundPlayers = append(finalRoundPlayers, player.Id)
		}
	}
	if len(finalRoundPlayers) == 0 {
		return false
	}

	r.FinalRoundState = &FinalRoundState{
		Players:             finalRoundPlayers,
		AvailableCategories: pack.FinalRound.getAvailableCategories(),
	}
	r.CurrentPlayer = &r.FinalRoundState.Players[rand.Intn(len(r.FinalRoundState.Players))]
	r.State = SelectingFinalRoundCategory

	availableFinalCategories := r.GetAvailableFinalRoundCategories()
	if len(availableFinalCategories) == 1 {
		r.chooseFinalRoundCategory(pack, availableFinalCategories[0])
	}
	return true
}

func (r *Room) GetAvailableFinalRoundCategories() []string {
	availableCategories := make([]string, 0)
	for category, available := range r.FinalRoundState.AvailableCategories {
		if available {
			availableCategories = append(availableCategories, category)
		}
	}
	return availableCategories
}

func (r *Room) RemoveFinalRoundCategory(pack *Pack, userId string, category string) error {
	if r.State != SelectingFinalRoundCategory {
		return custerr.NewConflictErr("cannot remove final round category now")
	}
	if userId != *r.CurrentPlayer && !r.IsUserHost(userId) {
		return custerr.NewForbiddenErr("not allowed to remove final round category")
	}

	available, ok := r.FinalRoundState.AvailableCategories[category]
	if !ok {
		return custerr.NewNotFoundErr(fmt.Sprintf("no final round category \"%s\"", category))
	}
	if !available {
		return custerr.NewConflictErr(fmt.Sprintf("final round category \"%s\" has already been removed", category))
	}

	r.FinalRoundState.AvailableCategories[category] = false
	playerIndex := slices.IndexFunc(r.FinalRoundState.Players, func(p string) bool {
		return p == userId
	})
	if playerIndex == len(r.FinalRoundState.Players)-1 {
		r.CurrentPlayer = &r.FinalRoundState.Players[0]
	} else {
		r.CurrentPlayer = &r.FinalRoundState.Players[playerIndex+1]
	}

	availableFinalCategories := r.GetAvailableFinalRoundCategories()
	if len(availableFinalCategories) == 1 {
		r.chooseFinalRoundCategory(pack, availableFinalCategories[0])
	}
	return nil
}

func (r *Room) chooseFinalRoundCategory(pack *Pack, category string) {
	categoryIndex := slices.IndexFunc(pack.FinalRound.Categories, func(frc FinalRoundCategory) bool {
		return frc.Name == category
	})
	r.FinalRoundState.Question = &pack.FinalRound.Categories[categoryIndex].Question
	r.CurrentPlayer = nil
	r.State = FinalRoundBetting
}

func (r *Room) PlaceFinalRoundBet(userId string, amount int) error {
	if r.State != FinalRoundBetting {
		return custerr.NewConflictErr("can not place final round bet now")
	}
	inFinalRound := slices.Contains(r.FinalRoundState.Players, userId)
	if !inFinalRound {
		return custerr.NewForbiddenErr("not allowed to place bet")
	}
	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return userId == p.Id
	})
	alreadyBet := r.Players[playerIndex].BetAmount != nil
	if alreadyBet {
		return custerr.NewConflictErr("can not place bet again")
	}
	// betIncreased := r.Players[playerIndex].BetAmount != nil && *r.Players[playerIndex].BetAmount < amount
	insufficientScore := amount > r.Players[playerIndex].Score || amount < 0
	if insufficientScore {
		return custerr.NewConflictErr("insufficient score")
	}

	r.Players[playerIndex].BetAmount = &amount

	allBet := !slices.ContainsFunc(r.Players, func(p Player) bool {
		return slices.Contains(r.FinalRoundState.Players, p.Id) && p.BetAmount == nil
	})
	if allBet {
		r.startFinalRoundQuestion()
	}
	return nil
}

func (r *Room) PlaceFinalRoundBetsAuto() {
	for i := range r.Players {
		if r.Players[i].BetAmount == nil {
			zero := 0
			r.Players[i].BetAmount = &zero
		}
	}
	r.startFinalRoundQuestion()
}

func (r *Room) startFinalRoundQuestion() {
	finalRoundPlayers := make([]string, 0)
	for _, p := range r.Players {
		if slices.Contains(r.FinalRoundState.Players, p.Id) && *p.BetAmount > 0 {
			finalRoundPlayers = append(finalRoundPlayers, p.Id)
		}
	}

	if len(finalRoundPlayers) == 0 {
		r.EndGame()
		return
	}

	r.FinalRoundState.Players = finalRoundPlayers
	r.AllowedToAnswer = finalRoundPlayers
	r.FinalRoundState.PlayersAnswers = make(map[string]string)

	duration := time.Duration(r.Options.QuestionThinkingTimeFinal) * time.Second
	timerEndsAt := time.Now().Add(duration)
	r.FinalRoundState.TimerEndsAt = &timerEndsAt
	r.State = ShowingFinalRoundQuestion
}

func (r *Room) SubmitFinalRoundAnswer(userId string, answer string) error {
	if r.State != ShowingFinalRoundQuestion {
		return custerr.NewConflictErr("can not submit final round answer now")
	}
	if !slices.Contains(r.AllowedToAnswer, userId) {
		return custerr.NewConflictErr("not allowed to submit final round answer")
	}

	r.FinalRoundState.PlayersAnswers[userId] = answer
	r.AllowedToAnswer = slices.DeleteFunc(r.AllowedToAnswer, func(playerId string) bool {
		return userId == playerId
	})

	if len(r.AllowedToAnswer) == 0 {
		r.EndFinalRoundQuestion()
	}

	return nil
}

func (r *Room) EndFinalRoundQuestion() {
	r.CurrentPlayer = &r.FinalRoundState.Players[0]
	r.State = ValidatingFinalRoundAnswers
}

func (r *Room) ValidateFinalRoundAnswer(userId string, isCorrect bool) error {
	if r.State != ValidatingFinalRoundAnswers {
		return custerr.NewConflictErr("can not validate final round answer now")
	}
	if !r.IsUserHost(userId) {
		return custerr.NewForbiddenErr("not allowed to validate final round answer")
	}

	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return p.Id == *r.CurrentPlayer
	})
	betAmount := *r.Players[playerIndex].BetAmount
	if isCorrect {
		r.Players[playerIndex].Score += betAmount
	} else {
		r.Players[playerIndex].Score -= betAmount
	}

	playerIndex = slices.IndexFunc(r.FinalRoundState.Players, func(p string) bool {
		return p == *r.CurrentPlayer
	})
	if playerIndex < len(r.FinalRoundState.Players)-1 {
		r.CurrentPlayer = &r.FinalRoundState.Players[playerIndex+1]
	} else {
		r.EndGame()
	}

	return nil
}

func (r *Room) EndGame() {
	// TODO: maybe some cleanup
	r.CurrentPlayer = nil
	if r.FinalRoundState != nil && r.FinalRoundState.TimerEndsAt != nil {
		r.FinalRoundState.TimerEndsAt = nil
	}
	r.State = GameOver
}
