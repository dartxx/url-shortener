package delete

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Alias string `json:"alias"`
}

type Response struct {
	response.Response
	Success bool `json:"success"`
}

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			log.Error("no alias in request")

			render.JSON(w, r, response.Error("no alias in request"))

			return
		}

		err = urlDeleter.DeleteURL(alias)

		if err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("alias not found", slog.String("alias", alias))

				render.JSON(w, r, response.Error("alias not found"))
			}

			log.Error("failed to delete url")
			sl.Err(err)

			render.JSON(w, r, response.Error("internal fail to delete url"))

			return
		}

		log.Info("url deleted", slog.String("alias", alias))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Success:  true,
		})
	}
}
