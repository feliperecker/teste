package core

import "net/http"

type Resource struct {
	Name       string
	Controller Controller
	Model      Model
	Prefix     string
	Path       string
}

func (r *Resource) BindRoutes(app App) error {
	addHTMLEndpoints := true
	enablePutUpdate := true
	// query:
	app.SetRoute("find_"+r.Name, &Route{
		Method:     http.MethodGet,
		Path:       r.Prefix + r.Path,
		Action:     r.Controller.Find,
		Template:   r.Name + "/query",
		Permission: "find_" + r.Name,
	})
	// findOne:
	app.SetRoute("findOne_"+r.Name, &Route{
		Method:     http.MethodGet,
		Path:       r.Prefix + r.Path + "/:id",
		Action:     r.Controller.FindOne,
		Template:   r.Name + "/findOne",
		Permission: "findOne_" + r.Name,
	})
	// create:
	app.SetRoute("create_"+r.Name, &Route{
		Method:     http.MethodPost,
		Path:       r.Prefix + r.Path,
		Action:     r.Controller.Create,
		Template:   r.Name + "/create",
		Permission: "create_" + r.Name,
	})
	if addHTMLEndpoints {
		app.SetRoute("create_page_"+r.Name, &Route{
			Method:     http.MethodGet,
			Path:       r.Prefix + r.Path,
			Action:     r.Controller.Create,
			Template:   r.Name + "/create",
			Permission: "create_" + r.Name,
		})
	}
	// update:
	app.SetRoute("update_"+r.Name, &Route{
		Method:     http.MethodPost,
		Path:       r.Prefix + r.Path + "/:id",
		Action:     r.Controller.Update,
		Template:   r.Name + "/update",
		Permission: "update_" + r.Name,
	})
	if enablePutUpdate {
		app.SetRoute("update_put_"+r.Name, &Route{
			Method:     http.MethodPut,
			Path:       r.Prefix + r.Path + "/:id",
			Action:     r.Controller.Update,
			Template:   r.Name + "/update",
			Permission: "update_" + r.Name,
		})
	}

	if addHTMLEndpoints {
		app.SetRoute("update_page_"+r.Name, &Route{
			Method:     http.MethodGet,
			Path:       r.Prefix + r.Path + "/:id",
			Action:     r.Controller.Update,
			Template:   r.Name + "/update",
			Permission: "update_" + r.Name,
		})
	}
	// delete
	app.SetRoute("delete_"+r.Name, &Route{
		Method:     http.MethodDelete,
		Path:       r.Prefix + r.Path + "/:id",
		Action:     r.Controller.Delete,
		Template:   r.Name + "/delete",
		Permission: "delete_" + r.Name,
	})
	// Count
	app.SetRoute("count_"+r.Name, &Route{
		Method:     http.MethodGet,
		Path:       r.Prefix + r.Path + "-count",
		Action:     r.Controller.Count,
		Permission: "find_" + r.Name,
	})

	return nil
}
