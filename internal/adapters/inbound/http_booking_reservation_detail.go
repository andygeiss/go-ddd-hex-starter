package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/redirecting"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// GuestInfoView represents guest information for the view.
type GuestInfoView struct {
	Name        string
	Email       string
	PhoneNumber string
}

// ReservationDetailView represents a reservation for the detail view.
type ReservationDetailView struct {
	ID                 string
	RoomID             string
	CheckIn            string
	CheckOut           string
	Status             string
	StatusClass        string
	TotalAmount        string
	CreatedAt          string
	CancellationReason string
	Guests             []GuestInfoView
	Nights             int
	CanCancel          bool
}

// HttpViewReservationDetailResponse specifies the view data for the reservation detail.
type HttpViewReservationDetailResponse struct {
	AppName     string
	Title       string
	SessionID   string
	Reservation ReservationDetailView
}

func buildReservationDetailView(reservation *booking.Reservation) ReservationDetailView {
	guests := make([]GuestInfoView, 0, len(reservation.Guests))
	for _, g := range reservation.Guests {
		guests = append(guests, GuestInfoView{
			Name:        g.Name,
			Email:       g.Email,
			PhoneNumber: g.PhoneNumber,
		})
	}

	return ReservationDetailView{
		Guests:             guests,
		ID:                 string(reservation.ID),
		RoomID:             string(reservation.RoomID),
		CheckIn:            reservation.DateRange.CheckIn.Format("2006-01-02"),
		CheckOut:           reservation.DateRange.CheckOut.Format("2006-01-02"),
		Status:             string(reservation.Status),
		StatusClass:        reservationStatusClass(reservation.Status),
		TotalAmount:        reservation.TotalAmount.FormatAmount(),
		CreatedAt:          reservation.CreatedAt.Format("2006-01-02 15:04"),
		CancellationReason: reservation.CancellationReason,
		Nights:             reservation.Nights(),
		CanCancel:          reservation.CanBeCancelled(),
	}
}

// HttpViewReservationDetail defines an HTTP handler function for rendering a single reservation.
func HttpViewReservationDetail(e *templating.Engine, reservationService *booking.ReservationService) http.HandlerFunc {
	appName := os.Getenv("APP_NAME")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionID, _ := ctx.Value(security.ContextSessionID).(string)
		email, _ := ctx.Value(security.ContextEmail).(string)
		if sessionID == "" || email == "" {
			redirecting.Redirect(w, r, "/ui/login")
			return
		}

		reservationID := r.PathValue("id")
		if reservationID == "" {
			http.Error(w, "Reservation ID required", http.StatusBadRequest)
			return
		}

		reservation, err := reservationService.GetReservation(ctx, booking.ReservationID(reservationID))
		if err != nil {
			http.Error(w, "Reservation not found", http.StatusNotFound)
			return
		}

		if string(reservation.GuestID) != email {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		data := HttpViewReservationDetailResponse{
			AppName:     appName,
			Title:       appName + " - Reservation " + reservationID,
			SessionID:   sessionID,
			Reservation: buildReservationDetailView(reservation),
		}

		HttpView(e, "reservation_detail", data)(w, r)
	}
}

// HttpCancelReservation handles the POST request to cancel a reservation.
func HttpCancelReservation(reservationService *booking.ReservationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check authentication
		sessionID, _ := ctx.Value(security.ContextSessionID).(string)
		email, _ := ctx.Value(security.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get reservation ID from path
		reservationID := r.PathValue("id")
		if reservationID == "" {
			http.Error(w, "Reservation ID required", http.StatusBadRequest)
			return
		}

		// Verify the reservation belongs to the current user
		reservation, err := reservationService.GetReservation(ctx, booking.ReservationID(reservationID))
		if err != nil {
			http.Error(w, "Reservation not found", http.StatusNotFound)
			return
		}

		if string(reservation.GuestID) != email {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Cancel the reservation
		err = reservationService.CancelReservation(ctx, booking.ReservationID(reservationID), "Cancelled by guest")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Redirect back to reservations list
		redirecting.Redirect(w, r, "/ui/reservations")
	}
}
