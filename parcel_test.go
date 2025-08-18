package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func openTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	return db
}

// TestAddGetDelete
func TestAddGetDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, stored.Client)
	require.Equal(t, parcel.Status, stored.Status)
	require.Equal(t, parcel.Address, stored.Address)
	require.Equal(t, parcel.CreatedAt, stored.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress
func TestSetAddress(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)

	// add
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, stored.Address)
}

// TestSetStatus
func TestSetStatus(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, stored.Status)
}

// TestGetByClient
func TestGetByClient(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for _, p := range storedParcels {
		expected, ok := parcelMap[p.Number]
		require.True(t, ok)
		require.Equal(t, expected.Client, p.Client)
		require.Equal(t, expected.Status, p.Status)
		require.Equal(t, expected.Address, p.Address)
		require.Equal(t, expected.CreatedAt, p.CreatedAt)
	}
}
