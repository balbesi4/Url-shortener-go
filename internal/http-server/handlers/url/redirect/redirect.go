package redirect

import (
	"Url-shortener-go/internal/lib/api/response"
	"Url-shortener-go/internal/lib/logger/slog_logger"
	"Url-shortener-go/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const f = "handlers.url.redirect.New"

		log = log.With(
			slog.String("function", f),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("redirect: missing alias")
			render.JSON(w, r, response.Error("missing alias"))
			return
		}

		url, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("redirect: url not found", "alias", alias)
			render.JSON(w, r, response.Error("url not found"))
			return
		}

		if err != nil {
			log.Info("redirect: error getting url", slog_logger.Err(err))
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
		log.Info("redirect: redirected to", "url", url)
	}
}
