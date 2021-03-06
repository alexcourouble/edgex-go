/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package notifications

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

const (
	applicationJson = "application/json; charset=utf-8"
)

func subscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	switch r.Method {

	// Get all subscriptions
	case http.MethodGet:
		subscriptions, err := dbClient.GetSubscriptions()
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		encode(subscriptions, w)

		// Modify (an existing) subscription
	case http.MethodPut:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		// Check if the subscription exists
		s2, err := dbClient.GetSubscriptionBySlug(s.Slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		} else {
			s.ID = s2.ID
		}

		LoggingClient.Info("Updating subscription by slug: " + slug)

		if err = dbClient.UpdateSubscription(s); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))

	case http.MethodPost:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding subscription: " + err.Error())
			return
		}

		LoggingClient.Info("Posting Subscription: " + s.String())
		_, err = dbClient.AddSubscription(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			LoggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(s.Slug))

	}
}

func subscriptionByIDHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	switch r.Method {
	case http.MethodGet:

		s, err := dbClient.GetSubscriptionById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		encode(s, w)
		break
	case http.MethodDelete:
		_, err := dbClient.GetSubscriptionById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Deleting subscription: " + id)

		if err = dbClient.DeleteSubscriptionById(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func subscriptionsBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	switch r.Method {
	case http.MethodGet:

		s, err := dbClient.GetSubscriptionBySlug(slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			encode(s, w)
			return
		}

		encode(s, w)
	case http.MethodDelete:
		_, err := dbClient.GetSubscriptionBySlug(slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Deleting subscription by slug: " + slug)

		if err = dbClient.DeleteSubscriptionBySlug(slug); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

func subscriptionsByCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	categories := splitVars(vars["categories"])

	s, err := dbClient.GetSubscriptionByCategories(categories)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	encode(s, w)

}

func subscriptionsByLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	labels := splitVars(vars["labels"])

	s, err := dbClient.GetSubscriptionByLabels(labels)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	encode(s, w)

}

func subscriptionsByCategoriesLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	labels := splitVars(vars["labels"])
	categories := splitVars(vars["categories"])

	s, err := dbClient.GetSubscriptionByCategoriesLabels(categories, labels)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	encode(s, w)

}

func splitVars(vars string) []string {
	return strings.Split(vars, ",")
}

func subscriptionsByReceiverHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	s, err := dbClient.GetSubscriptionByReceiver(vars["receiver"])
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}
	encode(s, w)

}
