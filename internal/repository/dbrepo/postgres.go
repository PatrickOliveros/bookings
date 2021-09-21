package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/patrickoliveros/bookings/models"
	"golang.org/x/crypto/bcrypt"
)

// region "Users"
func (m *postgresDBRepo) GetAllUsers() bool {
	return true
}

func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	query := `select id, first_name, last_name, email, password, access_level from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (m *postgresDBRepo) UpdateUser(u models.User) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $2, last_name = $3, 
		email = $4, password = $5, access_level = $6, updated_at = $7 where id = $1`

	_, err := m.DB.ExecContext(ctx, query, u.ID, u.FirstName, u.LastName, u.Email, u.Password, u.AccessLevel, time.Now())

	return err
}

func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword, firstName, lastName string

	query := `select id, password, first_name, last_name from users where email = $1`

	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword, &firstName, &lastName)

	if err != nil {
		return id, "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", "", err
	}

	return id, fmt.Sprintf("%s %s", firstName, lastName), hashedPassword, nil
}

// endregion

// region "Rooms"
func (m *postgresDBRepo) GetAllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		select rm.id, rm.room_name
			from rooms rm 
					order by room_name asc
		`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Room

		err := rows.Scan(
			&item.ID,
			&item.RoomName,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, item)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return room, err
	}

	return room, nil
}

// endregion

// region "Reservations"
func (m *postgresDBRepo) GetAllReservations(onlyNew bool) ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	var sb strings.Builder

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
		r.end_date, r.room_id, r.processed, r.reference, r.created_at, r.updated_at, 
		rm.id, rm.room_name
			from reservations r
				inner join rooms rm on r.room_id  = rm.id 
					
		`
	sb.WriteString(query)

	if onlyNew {
		sb.WriteString(`where r.processed = 0`)
	}

	sb.WriteString(`order by r.created_at asc`)

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Reservation

		err := rows.Scan(
			&item.ID,
			&item.FirstName,
			&item.LastName,
			&item.Email,
			&item.Phone,
			&item.StartDate,
			&item.EndDate,
			&item.RoomID,
			&item.Processed,
			&item.Reference,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Room.ID,
			&item.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, item)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservation models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
			r.end_date, r.room_id, r.processed, r.created_at, r.updated_at, r.reference,
				rm.id, rm.room_name
					from reservations r
						inner join rooms rm on r.room_id  = rm.id 
			 				where r.id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.Processed,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Reference,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
	)

	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	var newID int

	// we need to put things into context. if the process doesn't work within 3 seconds, something is wrong
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	stmt := `insert into reservations (first_name, last_name, email, phone,
		start_date, end_date, room_id, reference, created_at, updated_at) values
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id `

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName, res.LastName, res.Email, res.Phone,
		res.StartDate, res.EndDate, res.RoomID, res.Reference,
		time.Now(), time.Now()).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) UpdateReservation(r models.Reservation) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set first_name = $2, last_name = $3, 
		email = $4, phone = $5, updated_at = $6 where id = $1`

	_, err := m.DB.ExecContext(ctx, query,
		r.ID, r.FirstName, r.LastName, r.Email, r.Phone, time.Now())

	return err
}

func (m *postgresDBRepo) DeleteReservation(id int) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete from reservations where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)

	return err
}

func (m *postgresDBRepo) MarkProcessedReservation(id, processed int) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set processed = $1 where id = $2`

	_, err := m.DB.ExecContext(ctx, query, processed, id)

	return err
}

// endregion

// region "Room Restrictions"
func (m *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at) values
		($1, $2, $3, $4, $5, $6, $7) `

	_, err := m.DB.ExecContext(ctx, stmt, res.StartDate, res.EndDate, res.RoomID, res.ReservationID, res.RestrictionID, time.Now(), time.Now())

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `
		select id, reservation_id, restriction_id, room_id, start_date, end_date
		from room_restrictions where $1 < end_date and $2 >= start_date and room_id = $3`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.RoomRestriction

		err := rows.Scan(
			&item.ID,
			&item.ReservationID,
			&item.RestrictionID,
			&item.RoomID,
			&item.StartDate,
			&item.EndDate,
		)

		if err != nil {
			return restrictions, err
		}

		restrictions = append(restrictions, item)
	}

	if err = rows.Err(); err != nil {
		return restrictions, err
	}

	return restrictions, nil
}

// endregion

// region "Availability"
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoom(start, end time.Time, roomID int) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var availability int

	query := `select 
					count(id)
				from 
					room_restrictions rr 
				where 
					room_id = $1
					and $2 < end_date and $3 > start_date`

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&availability)

	if err != nil {
		log.Println(err)
		return false, err
	}

	if availability == 0 {
		return true, nil
	}

	return false, nil
}

func (m *postgresDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var rooms []models.Room

	query := `select 
					r.id, r.room_name
				from 
					rooms r
				where 
					r.id not in ( select rr.room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date )`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

//endregion

func (m *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, 
				restriction_id, reservation_id, created_at, updated_at) values
				($1, $2, $3, $4, $5, $6, $7) `

	_, err := m.DB.ExecContext(ctx, stmt,
		startDate, startDate.AddDate(0, 0, 1), id, 2, 0, time.Now(), time.Now())

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) DeleteBlocksForRoom(id int, blocks string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//stmt := `delete from room_restrictions where room_id = $1 and id in ($2)`
	stmt := fmt.Sprintf(`delete from room_restrictions where 
		room_id = %d and id in (%s)`, id, blocks)

	// log.Println(stmt2)

	_, err := m.DB.ExecContext(ctx, stmt)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
