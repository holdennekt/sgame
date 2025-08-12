package domain

type Event string

const (
	Chat                      Event = "chat"
	StartGame                 Event = "start_game"
	RoundStarted              Event = "round_started"
	RoundDemo                 Event = "round_demo"
	SelectQuestion            Event = "select_question"
	RevealingStarted          Event = "revealing_started"
	QuestionStarted           Event = "question_started"
	SubmitAnswer              Event = "submit_answer"
	PassingStarted            Event = "passing_started"
	PassQuestion              Event = "pass_question"
	BettingStarted            Event = "betting_started"
	PlaceBet                  Event = "place_bet"
	AnswerStarted             Event = "answer_started"
	ValidateAnswer            Event = "validate_answer"
	QuestionEnded             Event = "question_ended"
	CorrectAnswerDemo         Event = "correct_answer_demo"
	RemoveFinalRoundCategory  Event = "remove_final_round_category"
	FinalRoundBettingStarted  Event = "final_round_betting_started"
	PlaceFinalRoundBet        Event = "place_final_round_bet"
	FinalRoundQuestionStarted Event = "final_round_question_started"
	SubmitFinalRoundAnswer    Event = "submit_final_round_answer"
	ValidateFinalRoundAnswer  Event = "validate_final_round_answer"
	GameEnded                 Event = "game_ended"
	RoomUpdated               Event = "room_updated"
	RoomDeleted               Event = "room_deleted"
	UserDisconnected          Event = "user_disconnected"
	Error                     Event = "error"
)
