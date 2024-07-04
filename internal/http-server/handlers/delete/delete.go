package delete

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sqlite"
	"url-shortener/internal/storage"
)

// URLGetter is an interface for getting url by alias.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLDeleter interface {
	DeleteURL(alias string) (string, error)
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("Alias is empty")

			render.JSON(w, r, resp.Error("Invalid request"))

			return
		}

		resURL, err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("URL not found", "alias", alias)

			render.JSON(w, r, resp.Error("Not found"))

			return
		}
		if err != nil {
			log.Error("Failed to delete URL", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		log.Info("Got URL", slog.String("url", resURL))

		// redirect to found url
		http.NewRequest("DELETE", resURL, nil)
	}
}
