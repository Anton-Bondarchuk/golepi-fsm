package interfaces

import (
	"context"
	"errors"
	"sync"

	"golepi-fsm/internal/domain/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx context.Context, fsm *FSMContext, message *tgbotapi.Message, bot *models.Bot) error

type Router struct {
	Storage        Storage
	stateHandlers  map[models.State]HandlerFunc
	defaultHandler HandlerFunc
	mx sync.RWMutex
}

func NewRouter(storage Storage) *Router {
	return &Router{
		Storage:       storage,
		stateHandlers: make(map[models.State]HandlerFunc),
	}
}

func (r *Router) Register(state models.State, handler HandlerFunc) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.stateHandlers[state] = handler
}

func (r *Router) DefaultMessage(handler HandlerFunc) {
	r.defaultHandler = handler
}

func (r *Router) ProcessUpdate(ctx context.Context, update *tgbotapi.Update, bot *models.Bot) error {
	if update.Message == nil {
		return nil 
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	message := update.Message

	fsm := NewFSMContext(ctx, r.Storage, chatID, userID)

	state, err := fsm.Current()
	if err != nil {
		return err
	}

	r.mx.RLock()
	handler, exists := r.stateHandlers[state]
	defaultHandler := r.defaultHandler
	r.mx.RUnlock()

	
	if !exists {
		if defaultHandler != nil {
			return r.defaultHandler(ctx, fsm, message, bot)
		}
		return errors.New("no handler for state")
	}

	return handler(ctx, fsm, message, bot)
}
