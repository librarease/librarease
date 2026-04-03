package cli

import "strings"

func p(name string, kind ParamKind, required bool, help string) ParamSpec {
	return ParamSpec{Name: name, Flag: strings.ReplaceAll(name, "_", "-"), Kind: kind, Required: required, Help: help}
}

func endpointSpecs() map[string][]CommandSpec {
	return map[string][]CommandSpec{
		"system": {
			{Use: "hello", Short: "GET /api", Method: "GET", Path: "/api", ExpectEnvelope: false},
			{Use: "health", Short: "GET /api/health", Method: "GET", Path: "/api/health", ExpectEnvelope: false},
			{Use: "terms", Short: "GET /api/v1/terms", Method: "GET", Path: "/api/v1/terms", ExpectHTML: true, QueryParams: []ParamSpec{p("lang", ParamString, false, "Language code")}},
			{Use: "privacy", Short: "GET /api/v1/privacy", Method: "GET", Path: "/api/v1/privacy", ExpectHTML: true, QueryParams: []ParamSpec{p("lang", ParamString, false, "Language code")}},
		},
		"auth": {
			{Use: "register", Short: "POST /api/v1/auth/register", Method: "POST", Path: "/api/v1/auth/register", ExpectEnvelope: false, BodyParams: []ParamSpec{
				p("name", ParamString, true, "User name"),
				p("email", ParamString, true, "User email"),
				p("password", ParamString, true, "User password"),
			}},
		},
		"users": {
			{Use: "list", Short: "GET /api/v1/users", Method: "GET", Path: "/api/v1/users", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("skip", ParamInt, false, "Skip"),
				p("limit", ParamInt, false, "Limit"),
				p("sort_by", ParamString, false, "Sort by"),
				p("sort_in", ParamString, false, "Sort in"),
				p("name", ParamString, false, "Name"),
				p("global_role", ParamString, false, "Global role"),
				p("library_id", ParamString, false, "Library ID"),
			}},
			{Use: "get <id>", Short: "GET /api/v1/users/:id", Method: "GET", Path: "/api/v1/users/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("include_staffs", ParamBool, false, "Include staffs"),
			}},
			{Use: "create", Short: "POST /api/v1/users", Method: "POST", Path: "/api/v1/users", ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("name", ParamString, true, "Name"),
				p("email", ParamString, false, "Email"),
			}},
			{Use: "update <id>", Short: "PUT /api/v1/users/:id", Method: "PUT", Path: "/api/v1/users/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("name", ParamString, false, "Name"),
				p("phone", ParamString, false, "Phone"),
				p("global_role", ParamString, false, "Global role"),
			}},
			{Use: "delete <id>", Short: "DELETE /api/v1/users/:id", Method: "DELETE", Path: "/api/v1/users/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: false, NoContentSuccessText: "user deleted"},
			{Use: "me", Short: "GET /api/v1/users/me", Method: "GET", Path: "/api/v1/users/me", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("include", ParamSlice, false, "Include fields"),
			}},
		},
		"users_push_token": {
			{Use: "save", Short: "POST /api/v1/users/me/push-token", Method: "POST", Path: "/api/v1/users/me/push-token", ExpectEnvelope: false, NoContentSuccessText: "push token saved", BodyParams: []ParamSpec{
				p("token", ParamString, true, "Push token"),
				p("provider", ParamString, true, "Provider: fcm|apns|webpush"),
			}},
		},
		"users_watchlist": {
			{Use: "list", Short: "GET /api/v1/users/me/watchlist", Method: "GET", Path: "/api/v1/users/me/watchlist", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("skip", ParamInt, false, "Skip"),
				p("limit", ParamInt, false, "Limit"),
				p("library_id", ParamString, false, "Library ID"),
				p("title", ParamString, false, "Title"),
				p("sort_by", ParamString, false, "Sort by"),
				p("sort_in", ParamString, false, "Sort in"),
			}},
			{Use: "add", Short: "POST /api/v1/users/me/watchlist", Method: "POST", Path: "/api/v1/users/me/watchlist", ExpectEnvelope: false, BodyParams: []ParamSpec{
				p("book_id", ParamString, true, "Book ID"),
			}},
			{Use: "remove <book_id>", Short: "DELETE /api/v1/users/me/watchlist/:book_id", Method: "DELETE", Path: "/api/v1/users/me/watchlist/{book_id}", PathParamNames: []string{"book_id"}, ExpectEnvelope: false, NoContentSuccessText: "watchlist removed"},
		},
		"libraries": crudSpecs("/api/v1/libraries", []ParamSpec{
			p("name", ParamString, true, "Name"),
			p("logo", ParamString, false, "Logo"),
			p("address", ParamString, false, "Address"),
			p("phone", ParamString, false, "Phone"),
			p("email", ParamString, false, "Email"),
			p("description", ParamString, false, "Description"),
		}, []ParamSpec{
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("name", ParamString, false, "Name"),
		}),
		"staff": {
			{Use: "list", Short: "List resources", Method: "GET", Path: "/api/v1/staffs", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("skip", ParamInt, false, "Skip"),
				p("limit", ParamInt, false, "Limit"),
				p("sort_by", ParamString, false, "Sort by"),
				p("sort_in", ParamString, false, "Sort in"),
				p("library_id", ParamString, false, "Library ID"),
				p("user_id", ParamString, false, "User ID"),
				p("name", ParamString, false, "Name"),
				p("role", ParamString, false, "Role"),
			}},
			{Use: "get <id>", Short: "Get by ID", Method: "GET", Path: "/api/v1/staffs/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true},
			{Use: "create", Short: "Create resource", Method: "POST", Path: "/api/v1/staffs", ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("name", ParamString, true, "Name"),
				p("library_id", ParamString, true, "Library ID"),
				p("user_id", ParamString, true, "User ID"),
				{Name: "staff", Flag: "role", Kind: ParamString, Required: false, Help: "Role"},
			}},
			{Use: "update <id>", Short: "Update resource", Method: "PUT", Path: "/api/v1/staffs/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("name", ParamString, false, "Name"),
				p("role", ParamString, false, "Role"),
			}},
			{Use: "delete <id>", Short: "Delete resource", Method: "DELETE", Path: "/api/v1/staffs/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: false, NoContentSuccessText: "deleted"},
		},
		"memberships": crudSpecs("/api/v1/memberships", []ParamSpec{
			p("name", ParamString, true, "Name"),
			p("library_id", ParamString, true, "Library ID"),
			p("duration", ParamInt, true, "Duration"),
			p("active_loan_limit", ParamInt, true, "Active loan limit"),
			p("usage_limit", ParamInt, true, "Usage limit"),
			p("loan_period", ParamInt, true, "Loan period"),
			p("fine_per_day", ParamInt, false, "Fine per day"),
			p("price", ParamInt, false, "Price"),
			p("description", ParamString, false, "Description"),
		}, []ParamSpec{
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("name", ParamString, false, "Name"),
			p("library_id", ParamString, false, "Library ID"),
		}),
		"books": append(crudSpecs("/api/v1/books", []ParamSpec{
			p("title", ParamString, true, "Title"),
			p("author", ParamString, true, "Author"),
			p("year", ParamInt, true, "Year"),
			p("code", ParamString, true, "Code"),
			p("cover", ParamString, false, "Cover"),
			p("library_id", ParamString, true, "Library ID"),
			p("description", ParamString, false, "Description"),
		}, []ParamSpec{
			p("id", ParamString, false, "ID"),
			p("ids", ParamString, false, "CSV IDs"),
			p("library_id", ParamString, false, "Library ID"),
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("title", ParamString, false, "Title"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("include_stats", ParamBool, false, "Include stats"),
		}), CommandSpec{
			Use:            "import-preview",
			Short:          "GET /api/v1/books/import",
			Method:         "GET",
			Path:           "/api/v1/books/import",
			ExpectEnvelope: true,
			QueryParams: []ParamSpec{
				p("path", ParamString, true, "File path"),
				p("library_id", ParamString, true, "Library ID"),
			},
		}, CommandSpec{
			Use:            "import-confirm",
			Short:          "POST /api/v1/books/import",
			Method:         "POST",
			Path:           "/api/v1/books/import",
			ExpectEnvelope: true,
			BodyParams: []ParamSpec{
				p("path", ParamString, true, "File path"),
				p("library_id", ParamString, true, "Library ID"),
			},
		}),
		"subscriptions": crudSpecs("/api/v1/subscriptions", []ParamSpec{
			p("user_id", ParamString, true, "User ID"),
			p("membership_id", ParamString, true, "Membership ID"),
			p("note", ParamString, false, "Note"),
		}, []ParamSpec{
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("id", ParamString, false, "ID"),
			p("user_id", ParamString, false, "User ID"),
			p("membership_id", ParamString, false, "Membership ID"),
			p("library_id", ParamString, false, "Library ID"),
			p("membership_name", ParamString, false, "Membership name"),
			p("is_active", ParamBool, false, "Active only"),
			p("is_expired", ParamBool, false, "Expired only"),
		}),
		"borrowings": append(crudSpecs("/api/v1/borrowings", []ParamSpec{
			p("book_id", ParamString, true, "Book ID"),
			p("subscription_id", ParamString, true, "Subscription ID"),
			p("staff_id", ParamString, true, "Staff ID"),
			p("borrowed_at", ParamString, false, "Borrowed at RFC3339"),
			p("due_at", ParamString, false, "Due at RFC3339"),
			p("note", ParamString, false, "Note"),
		}, []ParamSpec{
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("book_id", ParamString, false, "Book ID"),
			p("subscription_id", ParamString, false, "Subscription ID"),
			p("membership_id", ParamString, false, "Membership ID"),
			p("library_id", ParamString, false, "Library ID"),
			p("user_id", ParamString, false, "User ID"),
			p("returning_id", ParamString, false, "Returning ID"),
			p("borrowed_at", ParamString, false, "Borrowed at RFC3339"),
			p("due_at", ParamString, false, "Due at RFC3339"),
			p("is_active", ParamBool, false, "Active"),
			p("is_overdue", ParamBool, false, "Overdue"),
			p("is_returned", ParamBool, false, "Returned"),
			p("is_lost", ParamBool, false, "Lost"),
			p("borrow_staff_id", ParamString, false, "Borrow staff ID"),
			p("return_staff_id", ParamString, false, "Return staff ID"),
			p("returned_at", ParamString, false, "Returned at RFC3339"),
			p("lost_at", ParamString, false, "Lost at RFC3339"),
			p("include_review", ParamBool, false, "Include review"),
		}), []CommandSpec{
			{Use: "return <id>", Short: "POST /api/v1/borrowings/:id/return", Method: "POST", Path: "/api/v1/borrowings/{id}/return", PathParamNames: []string{"id"}, ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("staff_id", ParamString, false, "Staff ID"),
				p("returned_at", ParamString, false, "Returned at RFC3339"),
				p("fine", ParamInt, false, "Fine"),
				p("note", ParamString, false, "Note"),
			}},
			{Use: "unreturn <id>", Short: "DELETE /api/v1/borrowings/:id/return", Method: "DELETE", Path: "/api/v1/borrowings/{id}/return", PathParamNames: []string{"id"}, ExpectEnvelope: true},
			{Use: "lost <id>", Short: "POST /api/v1/borrowings/:id/lost", Method: "POST", Path: "/api/v1/borrowings/{id}/lost", PathParamNames: []string{"id"}, ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("staff_id", ParamString, false, "Staff ID"),
				p("reported_at", ParamString, false, "Reported at RFC3339"),
				p("fine", ParamInt, false, "Fine"),
				p("note", ParamString, true, "Note"),
			}},
			{Use: "unlost <id>", Short: "DELETE /api/v1/borrowings/:id/lost", Method: "DELETE", Path: "/api/v1/borrowings/{id}/lost", PathParamNames: []string{"id"}, ExpectEnvelope: true},
			{Use: "export", Short: "POST /api/v1/borrowings/export", Method: "POST", Path: "/api/v1/borrowings/export", ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("library_id", ParamString, true, "Library ID"),
				p("is_active", ParamBool, false, "Active"),
				p("is_overdue", ParamBool, false, "Overdue"),
				p("is_returned", ParamBool, false, "Returned"),
				p("is_lost", ParamBool, false, "Lost"),
				p("borrowed_at_from", ParamString, false, "Borrowed at from RFC3339"),
				p("borrowed_at_to", ParamString, false, "Borrowed at to RFC3339"),
			}},
		}...),
		"analysis": {
			{Use: "summary", Short: "GET /api/v1/analysis", Method: "GET", Path: "/api/v1/analysis", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("from", ParamString, true, "From RFC3339"),
				p("to", ParamString, true, "To RFC3339"),
				p("limit", ParamInt, true, "Limit"),
				p("skip", ParamInt, false, "Skip"),
				p("library_id", ParamString, true, "Library ID"),
			}},
			{Use: "overdue", Short: "GET /api/v1/analysis/overdue", Method: "GET", Path: "/api/v1/analysis/overdue", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("from", ParamString, false, "From RFC3339"),
				p("to", ParamString, false, "To RFC3339"),
				p("library_id", ParamString, true, "Library ID"),
			}},
			{Use: "borrowing-heatmap", Short: "GET /api/v1/analysis/borrowing-heatmap", Method: "GET", Path: "/api/v1/analysis/borrowing-heatmap", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("start", ParamString, false, "Start RFC3339"),
				p("end", ParamString, false, "End RFC3339"),
				p("library_id", ParamString, true, "Library ID"),
			}},
			{Use: "returning-heatmap", Short: "GET /api/v1/analysis/returning-heatmap", Method: "GET", Path: "/api/v1/analysis/returning-heatmap", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("start", ParamString, false, "Start RFC3339"),
				p("end", ParamString, false, "End RFC3339"),
				p("library_id", ParamString, true, "Library ID"),
			}},
			{Use: "power-users", Short: "GET /api/v1/analysis/power-users", Method: "GET", Path: "/api/v1/analysis/power-users", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("from", ParamString, false, "From RFC3339"),
				p("to", ParamString, false, "To RFC3339"),
				p("library_id", ParamString, true, "Library ID"),
				p("limit", ParamInt, false, "Limit"),
				p("skip", ParamInt, false, "Skip"),
			}},
			{Use: "longest-unreturned", Short: "GET /api/v1/analysis/longest-unreturned", Method: "GET", Path: "/api/v1/analysis/longest-unreturned", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("from", ParamString, false, "From RFC3339"),
				p("to", ParamString, false, "To RFC3339"),
				p("library_id", ParamString, true, "Library ID"),
				p("limit", ParamInt, false, "Limit"),
				p("skip", ParamInt, false, "Skip"),
			}},
		},
		"files": {
			{Use: "upload-url", Short: "GET /api/v1/files/upload", Method: "GET", Path: "/api/v1/files/upload", ExpectEnvelope: false, QueryParams: []ParamSpec{
				p("name", ParamString, true, "File name"),
			}},
		},
		"notifications": {
			{Use: "list", Short: "GET /api/v1/notifications", Method: "GET", Path: "/api/v1/notifications", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("skip", ParamInt, false, "Skip"),
				p("limit", ParamInt, true, "Limit"),
				p("is_unread", ParamBool, false, "Unread only"),
			}},
			{Use: "create", Short: "POST /api/v1/notifications", Method: "POST", Path: "/api/v1/notifications", ExpectEnvelope: false, NoContentSuccessText: "notification created", BodyParams: []ParamSpec{
				p("user_id", ParamString, true, "User ID"),
				p("title", ParamString, true, "Title"),
				p("message", ParamString, true, "Message"),
				p("reference_id", ParamString, false, "Reference ID"),
				p("reference_type", ParamString, false, "Reference type"),
			}},
			{Use: "read <id>", Short: "POST /api/v1/notifications/:id/read", Method: "POST", Path: "/api/v1/notifications/{id}/read", PathParamNames: []string{"id"}, ExpectEnvelope: false, NoContentSuccessText: "notification marked read"},
			{Use: "read-all", Short: "POST /api/v1/notifications/read", Method: "POST", Path: "/api/v1/notifications/read", ExpectEnvelope: false, NoContentSuccessText: "all notifications marked read"},
		},
		"collections": {
			{Use: "list", Short: "GET /api/v1/collections", Method: "GET", Path: "/api/v1/collections", ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("library_id", ParamString, false, "Library ID"),
				p("title", ParamString, false, "Title"),
				p("book_title", ParamString, false, "Book title"),
				p("limit", ParamInt, false, "Limit"),
				p("offset", ParamInt, false, "Offset"),
				p("include_library", ParamBool, false, "Include library"),
				p("include_stats", ParamBool, false, "Include stats"),
				p("sort_by", ParamString, false, "Sort by"),
				p("sort_in", ParamString, false, "Sort in"),
			}},
			{Use: "create", Short: "POST /api/v1/collections", Method: "POST", Path: "/api/v1/collections", ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("library_id", ParamString, true, "Library ID"),
				p("title", ParamString, true, "Title"),
				p("cover", ParamString, false, "Cover"),
				p("description", ParamString, false, "Description"),
			}},
			{Use: "delete <id>", Short: "DELETE /api/v1/collections/:id", Method: "DELETE", Path: "/api/v1/collections/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: false},
			{Use: "get <id>", Short: "GET /api/v1/collections/:id", Method: "GET", Path: "/api/v1/collections/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("include_book_ids", ParamBool, false, "Include book ids"),
				p("include_stats", ParamBool, false, "Include stats"),
			}},
			{Use: "update <id>", Short: "PUT /api/v1/collections/:id", Method: "PUT", Path: "/api/v1/collections/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("title", ParamString, false, "Title"),
				p("description", ParamString, false, "Description"),
				p("update_cover", ParamString, false, "Update cover"),
			}},
			{Use: "books-list <collection_id>", Short: "GET /api/v1/collections/:collection_id/books", Method: "GET", Path: "/api/v1/collections/{collection_id}/books", PathParamNames: []string{"collection_id"}, ExpectEnvelope: true, QueryParams: []ParamSpec{
				p("include_book", ParamBool, false, "Include book"),
				p("book_title", ParamString, false, "Book title"),
				p("book_sort_by", ParamString, false, "Book sort by"),
				p("book_sort_in", ParamString, false, "Book sort in"),
				p("limit", ParamInt, false, "Limit"),
				p("skip", ParamInt, false, "Skip"),
				p("sort_by", ParamString, false, "Sort by"),
				p("sort_in", ParamString, false, "Sort in"),
			}},
			{Use: "books-update <collection_id>", Short: "PUT /api/v1/collections/:collection_id/books", Method: "PUT", Path: "/api/v1/collections/{collection_id}/books", PathParamNames: []string{"collection_id"}, ExpectEnvelope: true, BodyParams: []ParamSpec{
				p("book_ids", ParamSlice, false, "Book IDs"),
			}},
			{Use: "follow <collection_id>", Short: "POST /api/v1/collections/:collection_id/follow", Method: "POST", Path: "/api/v1/collections/{collection_id}/follow", PathParamNames: []string{"collection_id"}, ExpectEnvelope: false},
			{Use: "unfollow <collection_id>", Short: "DELETE /api/v1/collections/:collection_id/follow", Method: "DELETE", Path: "/api/v1/collections/{collection_id}/follow", PathParamNames: []string{"collection_id"}, ExpectEnvelope: false},
		},
		"jobs": append(crudReadOnlySpecs("/api/v1/jobs", []ParamSpec{
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("library_id", ParamString, true, "Library ID"),
			p("type", ParamString, false, "Type"),
			p("staff_id", ParamString, false, "Staff ID"),
			p("status", ParamString, false, "Status"),
		}), CommandSpec{
			Use:            "download <id>",
			Short:          "GET /api/v1/jobs/:id/download",
			Method:         "GET",
			Path:           "/api/v1/jobs/{id}/download",
			PathParamNames: []string{"id"},
			ExpectEnvelope: true,
		}),
		"reviews": crudSpecs("/api/v1/reviews", []ParamSpec{
			p("borrowing_id", ParamString, true, "Borrowing ID"),
			p("rating", ParamInt, true, "Rating 0-5"),
			p("comment", ParamString, false, "Comment"),
		}, []ParamSpec{
			p("skip", ParamInt, false, "Skip"),
			p("limit", ParamInt, false, "Limit"),
			p("sort_by", ParamString, false, "Sort by"),
			p("sort_in", ParamString, false, "Sort in"),
			p("borrowing_id", ParamString, false, "Borrowing ID"),
			p("user_id", ParamString, false, "User ID"),
			p("book_id", ParamString, false, "Book ID"),
			p("library_id", ParamString, false, "Library ID"),
			p("rating", ParamInt, false, "Rating"),
			p("comment", ParamString, false, "Comment"),
		}),
	}
}

func crudSpecs(base string, createBody []ParamSpec, listQuery []ParamSpec) []CommandSpec {
	return []CommandSpec{
		{Use: "list", Short: "List resources", Method: "GET", Path: base, ExpectEnvelope: true, QueryParams: listQuery},
		{Use: "get <id>", Short: "Get by ID", Method: "GET", Path: base + "/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true},
		{Use: "create", Short: "Create resource", Method: "POST", Path: base, ExpectEnvelope: true, BodyParams: createBody},
		{Use: "update <id>", Short: "Update resource", Method: "PUT", Path: base + "/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true, BodyParams: optionalize(createBody)},
		{Use: "delete <id>", Short: "Delete resource", Method: "DELETE", Path: base + "/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: false, NoContentSuccessText: "deleted"},
	}
}

func crudReadOnlySpecs(base string, listQuery []ParamSpec) []CommandSpec {
	return []CommandSpec{
		{Use: "list", Short: "List resources", Method: "GET", Path: base, ExpectEnvelope: true, QueryParams: listQuery},
		{Use: "get <id>", Short: "Get by ID", Method: "GET", Path: base + "/{id}", PathParamNames: []string{"id"}, ExpectEnvelope: true},
	}
}

func optionalize(in []ParamSpec) []ParamSpec {
	out := make([]ParamSpec, 0, len(in))
	for _, p := range in {
		p.Required = false
		out = append(out, p)
	}
	return out
}
