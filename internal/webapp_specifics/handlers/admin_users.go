package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func HandleAdminUserSearch(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminUserSearchResponse{Results: []webapp_models.AdminUserSearchResult{}})
			return
		}

		results := make([]webapp_models.AdminUserSearchResult, 0)

		if id, err := uuid.Parse(q); err == nil {
			user, err := cfg.DB.GetUserByIDAdmin(ctx, id)
			if err == nil {
				uid := user.ID.String()
				email := user.Email
				fn := user.FirstName
				ln := user.LastName
				ca := user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
				results = append(results, webapp_models.AdminUserSearchResult{
					UserID:    &uid,
					Email:     &email,
					FirstName: &fn,
					LastName:  &ln,
					CreatedAt: &ca,
				})
			} else {
				sessID := id
				sid := sessID.String()
				results = append(results, webapp_models.AdminUserSearchResult{
					SessionID: &sid,
				})
			}
		} else if strings.Contains(q, "@") {
			users, err := cfg.DB.SearchUsersByEmail(ctx, database.SearchUsersByEmailParams{
				Column1: sql.NullString{String: q, Valid: true},
				Limit:   20,
			})
			if err != nil {
				logger.Logger.Errorw("SearchUsersByEmail failed", "error", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Search failed")
				return
			}
			for _, u := range users {
				uid := u.ID.String()
				email := u.Email
				fn := u.FirstName
				ln := u.LastName
				ca := u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
				results = append(results, webapp_models.AdminUserSearchResult{
					UserID:    &uid,
					Email:     &email,
					FirstName: &fn,
					LastName:  &ln,
					CreatedAt: &ca,
				})
			}
		} else {
			users, err := cfg.DB.SearchUsersByEmail(ctx, database.SearchUsersByEmailParams{
				Column1: sql.NullString{String: q, Valid: true},
				Limit:   20,
			})
			if err == nil {
				for _, u := range users {
					uid := u.ID.String()
					email := u.Email
					fn := u.FirstName
					ln := u.LastName
					ca := u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
					results = append(results, webapp_models.AdminUserSearchResult{
						UserID:    &uid,
						Email:     &email,
						FirstName: &fn,
						LastName:  &ln,
						CreatedAt: &ca,
					})
				}
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminUserSearchResponse{Results: results})
	}
}

func HandleAdminUserDetail(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("id")
		userID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid user id")
			return
		}

		user, err := cfg.DB.GetUserByIDAdmin(ctx, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		snapRows, err := cfg.DB.ListFinancialSnapshotsForUser(ctx, database.ListFinancialSnapshotsForUserParams{
			UserID: uuid.NullUUID{UUID: userID, Valid: true},
			Limit:  12,
		})
		if err != nil {
			logger.Logger.Errorw("ListFinancialSnapshotsForUser failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load user")
			return
		}
		snapshots := make([]webapp_models.AdminSnapshotView, 0, len(snapRows))
		for _, snap := range snapRows {
			if v, ok := snapshotToAdminView(snap); ok {
				snapshots = append(snapshots, v)
			}
		}

		debtRows, err := cfg.DB.ListUserDebts(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("ListUserDebts failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load user")
			return
		}
		debts := make([]webapp_models.AdminDebtSummary, 0, len(debtRows))
		for _, d := range debtRows {
			item := webapp_models.AdminDebtSummary{
				ID:                 d.ID.String(),
				Name:               d.Name,
				Type:               string(d.Type),
				OutstandingBalance: d.OutstandingBalance.String(),
			}
			if d.TreatmentClassification.Valid {
				s := d.TreatmentClassification.String
				item.TreatmentClassification = &s
			}
			debts = append(debts, item)
		}

		fbRows, err := cfg.DB.ListFeedbackByUserID(ctx, database.ListFeedbackByUserIDParams{
			UserID: uuid.NullUUID{UUID: userID, Valid: true},
			Limit:  20,
		})
		if err != nil {
			logger.Logger.Errorw("ListFeedbackByUserID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load user")
			return
		}
		feedback := make([]webapp_models.AdminFeedbackRow, 0, len(fbRows))
		for _, row := range fbRows {
			feedback = append(feedback, mapFeedbackByUserRow(row))
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminUserDetailResponse{
			UserID:    user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			Snapshots: snapshots,
			Debts:     debts,
			Feedback:  feedback,
		})
	}
}

func HandleAdminSessionDetail(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("session_id")
		sessionID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid session id")
			return
		}

		snapRows, err := cfg.DB.ListFinancialSnapshotsBySession(ctx, database.ListFinancialSnapshotsBySessionParams{
			SessionID: uuid.NullUUID{UUID: sessionID, Valid: true},
			Limit:     12,
		})
		if err != nil {
			logger.Logger.Errorw("ListFinancialSnapshotsBySession failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load session")
			return
		}
		snapshots := make([]webapp_models.AdminSnapshotView, 0, len(snapRows))
		for _, snap := range snapRows {
			if v, ok := snapshotToAdminView(snap); ok {
				snapshots = append(snapshots, v)
			}
		}

		eventRows, err := cfg.DB.ListEventsBySession(ctx, sessionID)
		if err != nil {
			logger.Logger.Errorw("ListEventsBySession failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load session")
			return
		}
		events := make([]webapp_models.AdminCalculationEvent, 0, len(eventRows))
		for _, ev := range eventRows {
			item := webapp_models.AdminCalculationEvent{
				ID:        ev.ID.String(),
				EventType: ev.EventType,
				CreatedAt: ev.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			}
			if len(ev.Metadata) > 0 {
				var m map[string]interface{}
				if err := json.Unmarshal(ev.Metadata, &m); err == nil {
					item.Metadata = m
				}
			}
			events = append(events, item)
		}

		fbRows, err := cfg.DB.ListFeedbackBySessionID(ctx, database.ListFeedbackBySessionIDParams{
			SessionID: uuid.NullUUID{UUID: sessionID, Valid: true},
			Limit:     20,
		})
		if err != nil {
			logger.Logger.Errorw("ListFeedbackBySessionID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load session")
			return
		}
		feedback := make([]webapp_models.AdminFeedbackRow, 0, len(fbRows))
		for _, row := range fbRows {
			feedback = append(feedback, mapFeedbackBySessionRow(row))
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminSessionDetailResponse{
			SessionID: sessionID.String(),
			Snapshots: snapshots,
			Events:    events,
			Feedback:  feedback,
		})
	}
}
