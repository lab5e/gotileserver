package gotileserver

import (
	"embed" // For embedding the index page
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
)

//go:embed map/index.html
var indexPage []byte

//go:embed map/fonts
var fonts embed.FS

//go:embed map/styles
var styles embed.FS

//go:embed map/maplibre
var mapLibre embed.FS

type hostTemplate struct {
	Host string
}

// Assigned from external source
var tiles embed.FS

var overrideHost string = ""

var templates map[string]*template.Template

// RegisterHandler registers handlers for the /map path in the mux. Since the styles and spec files require
// a fair bit of massaging to work. The host override string
func RegisterHandler(mux *http.ServeMux, hostOverride string, tileSource embed.FS) error {
	overrideHost = hostOverride
	tiles = tileSource
	templates = make(map[string]*template.Template)

	var err error
	templates["bright.json"], err = template.New("bright.json").ParseFS(styles, "map/styles/bright.json")
	if err != nil {
		return err
	}
	templates["fiord.json"], err = template.New("fiord.json").ParseFS(styles, "map/styles/fiord.json")
	if err != nil {
		return err
	}
	templates["3d.json"], err = template.New("maptiler_3d.json").ParseFS(styles, "map/styles/3d.json")
	if err != nil {
		return err
	}
	templates["basic.json"], err = template.New("maptiler_basic.json").ParseFS(styles, "map/styles/basic.json")
	if err != nil {
		return err
	}
	templates["positron.json"], err = template.New("positron.json").ParseFS(styles, "map/styles/positron.json")
	if err != nil {
		return err
	}
	templates["metadata.json"], err = template.New("metadata.json").ParseFS(tiles, "map/tiles/metadata.json")
	if err != nil {
		return err
	}
	mux.HandleFunc("/map/index.html", indexHandler)
	mux.HandleFunc("/map/tiles/", tileHandler)

	mux.Handle("/map/maplibre/", http.FileServer(http.FS(mapLibre)))
	mux.Handle("/map/fonts/", http.FileServer(http.FS(fonts)))
	mux.HandleFunc("/map/styles/", styleHandler)

	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(indexPage)
}

func tileHandler(w http.ResponseWriter, r *http.Request) {
	var data []byte
	var err error
	fileName := r.URL.Path[1:]

	if strings.HasSuffix(fileName, ".json") {
		// This is the metadata file. Expand template and be happy
		tmpl, ok := templates["metadata.json"]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// This is the metadata; report appropriate content type
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := tmpl.Execute(w, getTemplateData(r)); err != nil {
			log.Printf("Error running metadata template: %v", err)
		}
		return
	}

	data, err = tiles.ReadFile(fileName)
	if err != nil {
		data, err = hex.DecodeString("1F8B0800FA78185E000393E2E3628F8F4FCD2D28A9D46850A86002006471443610000000")
		if err != nil {
			log.Printf("Empty tile error: %v\n", err)
		}
	}

	w.Header().Add("Content-Type", "application/x-protobuf")
	w.Header().Add("Content-Encoding", "gzip") // The generated tiles are gzipped by default
	w.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func styleHandler(w http.ResponseWriter, r *http.Request) {
	elems := strings.Split(r.URL.Path, "/")
	if len(elems) != 4 {
		// Check if this exists elsewhere in the file system
		buf, err := styles.ReadFile(r.URL.Path[1:])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Add("Content-Type", http.DetectContentType(buf))
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(buf)))
		w.Write(buf)
		return
	}
	templateName := elems[3]

	tmpl, ok := templates[templateName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if err := tmpl.ExecuteTemplate(w, templateName, getTemplateData(r)); err != nil {
		log.Printf("Error running style template: %v", err)
	}
}

// Make an attempt to guess the appropriate host name to use in the style templates. Because Reasons (tm) the
// styles need a complete URL for the resources and the front-end libraries gets *very* confused whent the
// URLs point to different places. Return the override setting if it is set (this is probably preferred if
// you have anything but the most trivial set-up)
func getTemplateData(r *http.Request) hostTemplate {
	if overrideHost != "" {
		return hostTemplate{
			Host: overrideHost,
		}
	}
	// Check if the origin header is set; use that one.
	host := r.Header.Get("Origin")
	if host != "" && host != "null" {
		return hostTemplate{
			Host: host,
		}
	}
	// This might be a direct request; check for Host + TLS config
	prefix := "http://"
	if r.TLS != nil {
		prefix = "https://"
	}
	return hostTemplate{
		Host: prefix + r.Host,
	}
}
