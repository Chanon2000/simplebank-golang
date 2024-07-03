package gapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/util"
	"github.com/chanon2000/simplebank/worker"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

