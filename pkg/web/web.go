package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/eu-evops/edulink/pkg/edulink"
	"github.com/eu-evops/edulink/pkg/util"
)

type Server struct {
	mux  *http.ServeMux
	port int
}

func NewServer(webserverPort int) *Server {
	return &Server{
		port: webserverPort,
	}
}

func (s *Server) Start() error {
	filePaths := []string{}
	filepath.WalkDir("site/templates", func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsDir() {
			return nil
		}

		if (filepath.Ext(path)) == ".tmpl" {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	templ := template.New("templates")
	templ.Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			json, _ := json.MarshalIndent(v, "", "  ")
			return string(json)
		},
	})
	templ = template.Must(templ.ParseFiles(filePaths...))

	s.mux = http.NewServeMux()

	s.mux.Handle(makeHandler("EduLink.SchoolDetails", makeEdulinkSchoolDetailsRequest, makeEdulinkSchoolDetailsResult, templ))
	s.mux.Handle(makeHandler("EduLink.AchievementBehaviourLookups", makeEdulinkAchievementBehaviourLookupsRequest, makeEdulinkAchievementBehaviourLookupsResult, templ))

	s.mux.Handle("/public/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.mux,
		ReadHeaderTimeout: 100 * time.Millisecond,
		WriteTimeout:      100 * time.Millisecond,
	}

	go server.ListenAndServe()

	defer server.Close()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigc
		log.Printf("Closing the HTTP server")
		server.Close()
	}()

	return nil
}

type LoggingHandler struct {
	handler http.HandlerFunc
}

type ResponseWriterContentLengthAware struct {
	http.ResponseWriter
	contentLength int
}

func (w *ResponseWriterContentLengthAware) Write(b []byte) (int, error) {
	w.contentLength += len(b)
	return w.ResponseWriter.Write(b)
}

func (h *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("Received request for %s - %s [%s]\n", r.URL.Path, r.Header.Get("user-agent"), r.RemoteAddr)

	timeout := 1510 * time.Millisecond

	wrappedHandler := http.TimeoutHandler(h.handler, timeout, "Timeout")
	wrapped := &ResponseWriterContentLengthAware{ResponseWriter: w}
	wrappedHandler.ServeHTTP(wrapped, r)

	end := time.Now()
	duration := end.Sub(start)
	log.Printf("Finished request for %s, duration: %dms, size: %dkb", r.URL.Path, duration.Milliseconds(), wrapped.contentLength/1/1024)
}

func makeEdulinkSchoolDetailsRequest(r *http.Request) edulink.Request {
	return &edulink.SchoolDetailsRequest{
		RequestBase: edulink.RequestBase{
			JsonRPC: "2.0",
			Method:  "EduLink.SchoolDetails",
		},
		Params: edulink.SchoolDetailsRequestParams{
			EstablishmentID: 2,
		},
	}
}

func makeEdulinkAchievementBehaviourLookupsRequest(r *http.Request) edulink.Request {
	return &edulink.AchievementBehaviourLookupsRequest{
		RequestBase: edulink.RequestBase{
			JsonRPC: "2.0",
			Method:  "EduLink.AchievementBehaviourLookups",
		},
	}
}

func makeEdulinkSchoolDetailsResult() edulink.Result {
	return &edulink.SchoolDetailsResponse{}
}

func makeEdulinkAchievementBehaviourLookupsResult() edulink.Result {
	return &edulink.AchievementBehaviourLookupsResponse{}
}

type makeEdulinkRequest func(*http.Request) edulink.Request
type makeEdulinkResult func() edulink.Result

func makeHandler(method string, makeRequest makeEdulinkRequest, makeResult makeEdulinkResult, templ *template.Template) (string, http.Handler) {
	h := func(w http.ResponseWriter, r *http.Request) {
		req := makeRequest(r)
		res := makeResult()
		if err := util.Call(r.Context(), req, res); err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		w.Header().Add("Content-Type", "text/html")

		log.Printf("Request to %s finished, rendering template\n", method)
		if err := templ.ExecuteTemplate(w, fmt.Sprintf("%s.go.tmpl", method), res); err != nil {
			log.Printf("Error: %s", err)
		}
	}

	return fmt.Sprintf("/%s", method), &LoggingHandler{handler: h}
}

func (s *Server) Stop() error {
	return nil
}
