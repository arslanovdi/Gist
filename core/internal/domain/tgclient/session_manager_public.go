package tgclient

import "context"

func (m *SessionManager) GetAllChats(ctx context.Context, userID int64) ([]string, error) {
	session, err := m.GetSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	chats, err := session.GetAllChats(ctx)
	if err != nil {
		return nil, err
	}
	return chats, nil
}
