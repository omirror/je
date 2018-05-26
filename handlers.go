package je

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	// Database
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"

	// Routing
	"github.com/julienschmidt/httprouter"
)

// SearchHandler ...
func (s *Server) SearchHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var jobs []*Job

		s.counters.Inc("n_search")

		qs := r.URL.Query()

		if id := SafeParseInt(p.ByName("id"), 0); id > 0 {
			err := db.Find("ID", id, &jobs)
			if err != nil && err == storm.ErrNotFound {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		} else if name := qs.Get("name"); name != "" {
			err := db.Select(q.Re("Name", name)).Find(&jobs)
			if err != nil && err == storm.ErrNotFound {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		} else if state := ParseState(qs.Get("state")); state != State(0) {
			err := db.Find("State", state, &jobs)
			if err != nil && err == storm.ErrNotFound {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		} else {
			err := db.All(&jobs)
			if err != nil {
				log.Errorf("error querying jobs index: %s", err)
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				return
			}
		}

		out, err := json.Marshal(jobs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/json")
		w.Write(out)
	}
}

// LogsHandler ...
func (s *Server) LogsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var job Job

		s.counters.Inc("n_logs")

		qs := r.URL.Query()
		id := SafeParseInt(p.ByName("id"), 0)

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &job)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if qs.Get("follow") == "" {
			logs, err := job.Logs()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer logs.Close()

			w.Header().Set("Context-Type", "text/plaino")
			io.Copy(w, logs)
		} else {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lines, errors := job.LogsTail(ctx)
			for {
				select {
				case line := <-lines:
					message := []byte(fmt.Sprintf("%s\n", line))
					// TODO: What if n < len(message)?
					_, err = w.Write(message)
					// TODO: Resend?
					if err != nil {
						log.Errorf("error streaming output for job #%d: %s", job.ID, err)
						return
					}

					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					} else {
						log.Warn("no flusher")
					}
				case err := <-errors:
					log.Errorf("error reading output for job #%d: %s", job.ID, err)
					return
				}
			}
		}
	}
}

// OutputHandler ...
func (s *Server) OutputHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var job Job

		s.counters.Inc("n_output")

		qs := r.URL.Query()
		id := SafeParseInt(p.ByName("id"), 0)

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &job)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if qs.Get("follow") == "" {
			output, err := job.Output()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer output.Close()

			w.Header().Set("Context-Type", "text/plaino")
			io.Copy(w, output)
		} else {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lines, errors := job.OutputTail(ctx)
			for {
				select {
				case line := <-lines:
					message := []byte(fmt.Sprintf("%s\n", line))
					// TODO: What if n < len(message)?
					_, err = w.Write(message)
					// TODO: Resend?
					if err != nil {
						log.Errorf("error streaming output for job #%d: %s", job.ID, err)
						return
					}

					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					} else {
						log.Warn("no flusher")
					}
				case err := <-errors:
					log.Errorf("error reading output for job #%d: %s", job.ID, err)
					return
				}
			}
		}
	}
}

// KillHandler ...
func (s *Server) KillHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var job Job

		s.counters.Inc("n_kill")

		qs := r.URL.Query()
		id := SafeParseInt(p.ByName("id"), 0)

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &job)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		worker := s.pool.GetWorker(job.Worker)
		if worker == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = worker.Kill(qs.Get("force") == "")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// CreateHandler ...
func (s *Server) CreateHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_create")

		qs := r.URL.Query()

		name := p.ByName("name")
		if name == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		args := strings.Fields(qs.Get("args"))

		job, err := NewJob(name, args)
		if err != nil {
			log.Errorf("error creating new job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		err = job.SetInput(r.Body)
		if err != nil {
			log.Errorf("error setting job input for #%d: %s", job.ID, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		err = s.pool.Submit(job)
		if err != nil {
			log.Errorf("error submitting job to pool: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		if qs.Get("wait") != "" {
			job.Wait()
		}

		u, err := url.Parse(fmt.Sprintf("/search/%d", job.ID))
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		http.Redirect(w, r, r.URL.ResolveReference(u).String(), http.StatusFound)
	}
}

// StatsHandler ...
func (s *Server) StatsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bs, err := json.Marshal(s.stats.Data())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
	}
}
