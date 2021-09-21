package repository

import (
	"time"

	"github.com/patrickoliveros/bookings/models"
)

type DatabaseRepo interface {

	// Users
	GetAllUsers() bool
	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, string, error)

	// Rooms
	GetAllRooms() ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)

	// Reservations
	GetAllReservations(onlyNew bool) ([]models.Reservation, error)
	GetReservationById(id int) (models.Reservation, error)
	InsertReservation(res models.Reservation) (int, error)
	UpdateReservation(res models.Reservation) error
	DeleteReservation(id int) error
	MarkProcessedReservation(id, processed int) error

	// Room Restrictions
	InsertRoomRestriction(res models.RoomRestriction) error
	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlocksForRoom(id int, blocks string) error

	// Availability
	SearchAvailabilityByDatesByRoom(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error)
}
