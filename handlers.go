package amznode

import (
	"net/http"
)

func (s *server) createHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parentID, err := urlParamID(r, "parentID")
		if err != nil {
			respondErr(w, r, err, http.StatusBadRequest)
			return
		}
		childName, err := urlParamName(r, "childName")
		if err != nil {
			respondErr(w, r, err, http.StatusBadRequest)
			return
		}

		child, err := s.storage.Create(childName, parentID)
		if err != nil {
			handleStorageError(w, r, err)
			return
		}

		respond(w, r, child, http.StatusCreated)
	}
}

func (s *server) getHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := urlParamID(r, "id")
		if err != nil {
			respondErr(w, r, err, http.StatusBadRequest)
			return
		}
		if id == 0 {
			roots, err := s.storage.GetRoots()
			if err != nil {
				handleStorageError(w, r, err)
				return
			}
			respond(w, r, roots, http.StatusOK)
			return
		}
		node, err := s.storage.Get(id)
		if err != nil {
			handleStorageError(w, r, err)
			return
		}

		respond(w, r, node, http.StatusOK)
	}
}

func (s *server) changeParentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := urlParamID(r, "id")
		if err != nil {
			respondErr(w, r, err, http.StatusBadRequest)
			return
		}
		parentID, err := urlParamID(r, "parentID")
		if err != nil {
			respondErr(w, r, err, http.StatusBadRequest)
			return
		}

		err = s.storage.ChangeParent(id, parentID)
		if err != nil {
			handleStorageError(w, r, err)
			return
		}

		respond(w, r, nil, http.StatusOK)
	}
}

func (s *server) deleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := urlParamID(r, "id")
		if err != nil {
			respondErr(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.storage.Delete(id); err != nil {
			handleStorageError(w, r, err)
			return
		}

		respond(w, r, nil, http.StatusOK)
	}
}
