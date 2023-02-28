package server

import "golang.org/x/exp/slog"

type ServerOpt func(svr *StockServer)

func WithHost(host string) ServerOpt {
	return func(svr *StockServer) {
		svr.host = host
	}
}

func WithPort(port int) ServerOpt {
	return func(svr *StockServer) {
		svr.port = port
	}
}

func WithLogger(logger *slog.Logger) ServerOpt {
	return func(svr *StockServer) {
		svr.logger = logger.With(
			slog.Group("server_info",
				slog.String("host", svr.host),
				slog.Int("port", svr.port),
			),
		)
	}
}
