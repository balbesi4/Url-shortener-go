package save

import (
	"Url-shortener-go/internal/lib/api/response"
	"Url-shortener-go/internal/lib/logger/slog_logger"
	"Url-shortener-go/internal/lib/random"
	"Url-shortener-go/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

const aliasLength = 8

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(fullURL string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const f = "handlers.url.save.New"

		log = log.With(
			slog.String("function", f),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var request Request

		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Error("failed to decode JSON", slog_logger.Err(err))
			render.JSON(w, r, response.Error("failed to decode JSON"))
			return
		}

		log.Info("request body decoded", slog.Any("request", request))

		if err = validator.New().Struct(request); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("failed to validate request", slog_logger.Err(err))
			render.JSON(w, r, response.Validation(validationErrors))
			return
		}

		alias := request.Alias
		if alias == "" {
			alias = random.GenerateRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(request.Url, alias)
		if errors.Is(err, storage.ErrUrlExists) {
			log.Info("url already exists", slog.String("url", request.Url))
			render.JSON(w, r, response.Error("url already exists"))
			return
		}

		if err != nil {
			log.Error("failed to save url", slog_logger.Err(err))
			render.JSON(w, r, response.Error("failed to save url"))
			return
		}

		log.Info(
			"url saved",
			slog.String("url", request.Url),
			slog.Int64("id", id),
		)

		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
