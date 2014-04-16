package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// AlbumsResponse represents the JSON response for /api/albums
type AlbumsResponse struct {
	Error  *Error       `json:"error"`
	Albums []data.Album `json:"albums"`
	Songs  []data.Song  `json:"songs"`
}

// GetAlbums retrieves one or more albums from wavepipe, and returns a HTTP status and JSON
func GetAlbums(r render.Render, params martini.Params) {
	// Output struct for albums request
	res := AlbumsResponse{}

	// List of albums to return
	albums := make([]data.Album, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "unsupported API version: " + version
			r.JSON(400, res)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "invalid integer album ID"
			r.JSON(400, res)
			return
		}

		// Load the album
		album := new(data.Album)
		album.ID = id
		if err := album.Load(); err != nil {
			res.Error = new(Error)

			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.Error.Code = 404
				res.Error.Message = "album ID not found"
				r.JSON(404, res)
				return
			}

			// All other errors
			log.Println(err)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// On single album, load the songs for this album
		songs, err := data.DB.SongsForAlbum(album.ID)
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Add songs to output
		res.Songs = songs

		// Add album to slice
		albums = append(albums, *album)
	} else {
		// Retrieve all albums
		tempAlbums, err := data.DB.AllAlbums()
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Copy albums into the output slice
		albums = tempAlbums
	}

	// Build response
	res.Error = nil
	res.Albums = albums

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}
