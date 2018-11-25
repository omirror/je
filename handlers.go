package je

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	// Routing
	"github.com/julienschmidt/httprouter"
)

// IndexHandler ...
func (s *Server) IndexHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		metrics.CounterVec("server", "requests").WithLabelValues("GET", "/").Inc()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Job Engine %s", FullVersion())
	}
}

// SearchHandler ...
func (s *Server) SearchHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var (
			err  error
			jobs []*Job
		)

		metrics.CounterVec("server", "requests").WithLabelValues("GET", "/search").Inc()

		qs := r.URL.Query()

		if id := ParseId(p.ByName("id")); id > 0 {
			jobs, err = store.Find(id)
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		} else if q := qs.Get("q"); q != "" {
			jobs, err = store.Search(q)
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		} else {
			jobs, err = store.All()
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		}

		out, err := json.Marshal(jobs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}
}

// LogsHandler ...
func (s *Server) LogsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		metrics.CounterVec("server", "requests").WithLabelValues("GET", "/logs").Inc()

		qs := r.URL.Query()
		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if qs.Get("follow") == "" {
			logs, err := data.Read(job.ID, DATA_LOGS)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer logs.Close()

			w.Header().Set("Content-Type", "text/plain")
			io.Copy(w, logs)
		} else {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lines, errors := data.Tail(job.ID, DATA_LOGS, ctx)
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
		metrics.CounterVec("server", "requests").WithLabelValues("GET", "/output").Inc()

		qs := r.URL.Query()
		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if qs.Get("follow") == "" {
			output, err := data.Read(job.ID, DATA_OUTPUT)
			if err != nil {
				log.Errorf("error reading job output for #%d: %s", id, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer output.Close()

			w.Header().Set("Content-Type", "text/plain")
			io.Copy(w, output)
		} else {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lines, errors := data.Tail(job.ID, DATA_OUTPUT, ctx)
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
		metrics.CounterVec("server", "requests").WithLabelValues("POST", "/kill").Inc()

		qs := r.URL.Query()
		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		worker := s.GetWorker(job.Worker)
		if worker == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = worker.Kill(qs.Get("force") != "")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// CreateHandler ...
func (s *Server) CreateHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		metrics.CounterVec("server", "requests").WithLabelValues("POST", "/create").Inc()

		qs := r.URL.Query()

		name := strings.TrimPrefix(p.ByName("name"), "/")
		if name == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		args := strings.Fields(qs.Get("args"))
		interactive := qs.Get("interactive") != ""

		job, err := NewJob(name, args, interactive)
		if err != nil {
			log.Errorf("error creating new job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		input, err := data.Write(job.ID, DATA_INPUT)
		if err != nil {
			log.Errorf("error creating job input for #%d: %s", job.ID, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		n, err := io.Copy(input, r.Body)
		log.Debugf("written %d bytes of input for job #%d", n, job.ID)
		if cl := r.ContentLength; cl != 0 && cl < n {
			// TODO: Bump a counter?
			log.Warnf("not all bytes %d/%d of input written to job #%d", n, cl, job.ID)
		}
		if err != nil {
			log.Errorf("error writing input for job #%d: %s", job.ID, err)
		}

		err = input.Close()
		if err != nil {
			log.Errorf("error closing input for job #%d: %s", job.ID, err)
		}

		err = s.queue.Publish(job)
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

// ReadHandler ...
func (s *Server) ReadHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		metrics.CounterVec("server", "requests").WithLabelValues("GET", "/read").Inc()

		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		output, err := data.Read(job.ID, DATA_INPUT)
		if err != nil {
			log.Errorf("error reading job input for #%d: %s", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer output.Close()

		w.Header().Set("Content-Type", "text/plain")
		io.Copy(w, output)
	}
}

// WriteHandler ...
func (s *Server) WriteHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		metrics.CounterVec("server", "requests").WithLabelValues("POST", "/write").Inc()

		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		worker := s.GetWorker(job.Worker)
		if worker == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// TODO: WHat if n < len(r.Body)?
		_, err = worker.Write(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// CloseHandler ...
func (s *Server) CloseHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		metrics.CounterVec("server", "requests").WithLabelValues("POST", "/close").Inc()

		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		worker := s.GetWorker(job.Worker)
		if worker == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = worker.Close()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// UpdateHandler ...
func (s *Server) UpdateHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var job Job

		metrics.CounterVec("server", "requests").WithLabelValues("POST", "/update").Inc()

		id := ParseId(p.ByName("id"))

		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		_, err := store.Get(id)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("error reading job update request body for #%d: %s", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(body, &job)
		if err != nil {
			log.Errorf("error decoding job update request for #%d: %s", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = store.Save(&job)
		if err != nil {
			log.Errorf("error saving job #%d: %s", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
