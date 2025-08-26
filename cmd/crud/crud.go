package crud

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

type TemplateData struct {
	PascalCase string
	CamelCase  string
	LowerCase  string
	KebabCase  string
}

var rootCmd = &cobra.Command{
	Use:   "gocrud-gen",
	Short: "A CLI tool to generate CRUD boilerplate for Go projects.",
	Long: `gocrud-gen is a command-line tool that automates the creation of 
repository, service, and controller layers for a new entity.`,
}

var crudCmd = &cobra.Command{
	Use:   "crud [EntityName]",
	Short: "Generates a full CRUD flow (repository, service, controller) for a new entity.",
	Long: `This command automates the creation of boilerplate files for a new entity.
You must provide the entity name in PascalCase. For example:

go run . crud SbsFee`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entityName := args[0]
		generateCrud(entityName)
	},
}

func init() {
	rootCmd.AddCommand(crudCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred: '%s'", err)
		os.Exit(1)
	}
}

func toKebabCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(s, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}

func generateCrud(namePascal string) {
	fmt.Printf("--- Generating CRUD for entity: %s ---\n", namePascal)

	data := TemplateData{
		PascalCase: namePascal,
		CamelCase:  strings.ToLower(namePascal[:1]) + namePascal[1:],
		LowerCase:  strings.ToLower(namePascal),
		KebabCase:  toKebabCase(namePascal),
	}

	filesToGenerate := map[string]string{
		filepath.Join("internal/transport/repository/postgres", data.CamelCase+".go"):                repositoryTemplate,
		filepath.Join("internal/service", data.CamelCase+".go"):                                      serviceTemplate,
		filepath.Join("internal/transport/http/rest/controller/v1", data.CamelCase, "controller.go"): controllerTemplate,
		filepath.Join("internal/transport/http/rest/controller/v1", data.CamelCase, "request.go"):    requestTemplate,
	}

	for path, tmplStr := range filesToGenerate {

		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Skipping existing file: %s.\n", path)
			continue
		} else if !os.IsNotExist(err) {
			fmt.Printf("Error checking file status for %s: %v\n", path, err)
			return
		} else {
			fmt.Printf("Generating file: %s\n", path)
		}

		// Create directories if they don't exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}

		file, err := os.Create(path)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", path, err)
			return
		}
		defer file.Close()

		tmpl, err := template.New(path).Parse(tmplStr)
		if err != nil {
			fmt.Printf("Error parsing template for %s: %v\n", path, err)
			return
		}
		if err := tmpl.Execute(file, data); err != nil {
			fmt.Printf("Error executing template for %s: %v\n", path, err)
			return
		}
	}

	fmt.Println("--- CRUD for", data.PascalCase, "generated successfully! ---")
	fmt.Println("Next steps:")
	fmt.Println("1. Define the 'dto.", data.PascalCase, "' struct in a relevant DTO file and ensure it implements 'dto.Entity'.")
	fmt.Printf("2. Populate the request structs in '%s'.\n", filepath.Join("internal/transport/http/rest/controller/v1", data.LowerCase, "request.go"))
	fmt.Println("3. Implement the TODOs in the generated controller to map request structs to your DTO.")
	fmt.Println("4. Add the new controller, service, and repository to the initializers in 'internal/initializer/app.go'.")
	fmt.Println("5. Add the new routes to the router in 'internal/transport/http/rest/router/route.go'.")
	fmt.Println("6. Update the ColumnMapping in the generated controller for filtering and sorting.")
}

// --- TEMPLATES ---

const requestTemplate = `package {{.LowerCase}}

type create{{.PascalCase}}Request struct {
	// TODO: Add fields for creating a new {{.PascalCase}}.
	// Example:
	// Name string ` + "`json:\"name\" validate:\"required\"`" + `
}

type update{{.PascalCase}}Request struct {
	// TODO: Add fields for updating an existing {{.PascalCase}}.
	// Example:
	// Name string ` + "`json:\"name\" validate:\"required\"`" + `
}
`

const repositoryTemplate = `package postgres

import (
	"git.snapp.ninja/search-and-discovery/framework/pkg/ports"
	dto "git.snapp.ninja/snappshop/delivery/harley/internal/DTO"
	"git.snapp.ninja/snappshop/delivery/harley/internal/transport/repository"
)

type {{.CamelCase}}Repository struct {
	repository.GenericRepository[dto.{{.PascalCase}}]
	db  ports.Database
	log ports.LoggerWithTraceID
}

func New{{.PascalCase}}Repository(db ports.Database, log ports.LoggerWithTraceID) repository.{{.PascalCase}} {
	return &{{.CamelCase}}Repository{
		GenericRepository: repository.NewGenericRepository[dto.{{.PascalCase}}](db, log),
		db:                db,
		log:               log,
	}
}
`

const serviceTemplate = `package service

import (
	"context"

	"git.snapp.ninja/search-and-discovery/framework/pkg/ports"
	dto "git.snapp.ninja/snappshop/delivery/harley/internal/DTO"
	"git.snapp.ninja/snappshop/delivery/harley/internal/transport/repository"
)

type {{.PascalCase}} interface {
	Get{{.PascalCase}}ByID(ctx context.Context, id int64) (dto.{{.PascalCase}}, error)
	Update{{.PascalCase}}(ctx context.Context, {{.CamelCase}} dto.{{.PascalCase}}) (dto.{{.PascalCase}}, error)
	Create{{.PascalCase}}(ctx context.Context, {{.CamelCase}} dto.{{.PascalCase}}) (dto.{{.PascalCase}}, error)
	Delete{{.PascalCase}}(ctx context.Context, id int64) error
	GetPaginated{{.PascalCase}}s(ctx context.Context, pagination dto.Pagination) ([]dto.{{.PascalCase}}, *dto.Pagination, error)
}

type {{.CamelCase}}Service struct {
	log              ports.LoggerWithTraceID
	{{.CamelCase}}Repository repository.{{.PascalCase}}
}

func New{{.PascalCase}}Service(log ports.LoggerWithTraceID, {{.CamelCase}}Repository repository.{{.PascalCase}}) {{.PascalCase}} {
	return &{{.CamelCase}}Service{
		log:              log,
		{{.CamelCase}}Repository: {{.CamelCase}}Repository,
	}
}

func (s *{{.CamelCase}}Service) Get{{.PascalCase}}ByID(ctx context.Context, id int64) (dto.{{.PascalCase}}, error) {
	{{.CamelCase}}, err := s.{{.CamelCase}}Repository.GetByID(ctx, id)
	if err != nil {
		return dto.{{.PascalCase}}{}, err
	}
	return {{.CamelCase}}, nil
}

func (s *{{.CamelCase}}Service) Update{{.PascalCase}}(ctx context.Context, {{.CamelCase}} dto.{{.PascalCase}}) (dto.{{.PascalCase}}, error) {
	err := s.{{.CamelCase}}Repository.Update(ctx, &{{.CamelCase}})
	if err != nil {
		return dto.{{.PascalCase}}{}, err
	}
	return {{.CamelCase}}, nil
}

func (s *{{.CamelCase}}Service) Create{{.PascalCase}}(ctx context.Context, {{.CamelCase}} dto.{{.PascalCase}}) (dto.{{.PascalCase}}, error) {
	err := s.{{.CamelCase}}Repository.Create(ctx, &{{.CamelCase}})
	if err != nil {
		return dto.{{.PascalCase}}{}, err
	}
	return {{.CamelCase}}, nil
}

func (s *{{.CamelCase}}Service) Delete{{.PascalCase}}(ctx context.Context, id int64) error {
	err := s.{{.CamelCase}}Repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *{{.CamelCase}}Service) GetPaginated{{.PascalCase}}s(ctx context.Context, pagination dto.Pagination) ([]dto.{{.PascalCase}}, *dto.Pagination, error) {
	{{.CamelCase}}s, resultPagination, err := s.{{.CamelCase}}Repository.FindAll(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}
	return {{.CamelCase}}s, resultPagination, nil
}
`

const controllerTemplate = `package {{.LowerCase}}

import (
	"errors"

	"git.snapp.ninja/search-and-discovery/framework/pkg/adapters/errorUtil/appErr"
	"git.snapp.ninja/search-and-discovery/framework/pkg/ports"
	dto "git.snapp.ninja/snappshop/delivery/harley/internal/DTO"
	"git.snapp.ninja/snappshop/delivery/harley/internal/consts"
	"git.snapp.ninja/snappshop/delivery/harley/internal/service"
	"git.snapp.ninja/snappshop/delivery/harley/internal/transport/http/rest/httpUtils"
	"git.snapp.ninja/snappshop/delivery/harley/internal/transport/http/rest/validator"
	"git.snapp.ninja/snappshop/delivery/harley/internal/utils"
	"go.elastic.co/apm"
)

type {{.PascalCase}} interface {
	GetPaginated{{.PascalCase}}s(c *ports.HttpContext) error
	Create{{.PascalCase}}(c *ports.HttpContext) error
	Get{{.PascalCase}}ByID(c *ports.HttpContext) error
	Update{{.PascalCase}}(c *ports.HttpContext) error
	Delete{{.PascalCase}}(c *ports.HttpContext) error
}

type {{.CamelCase}}Controller struct {
	{{.CamelCase}}Service    service.{{.PascalCase}}
	customValidation validator.CustomValidation
	log              ports.LoggerWithTraceID
}

func New(log ports.LoggerWithTraceID, {{.CamelCase}}Service service.{{.PascalCase}}, customValidation validator.CustomValidation) {{.PascalCase}} {
	return &{{.CamelCase}}Controller{
		{{.CamelCase}}Service:    {{.CamelCase}}Service,
		customValidation: customValidation,
		log:              log,
	}
}

// @Summary		Create a {{.PascalCase}}
// @Description	This route will create a {{.LowerCase}}
// @Tags			{{.PascalCase}}
// @Accept			json
// @Produce		json
// @Param			body	body		create{{.PascalCase}}Request 	true	"Create {{.PascalCase}} request"
// @Success		201		{object}	ports.Response{data=dto.{{.PascalCase}}}
// @Failure		400		{object}	ports.ErrorDetails
// @Failure		422		{object}	ports.ErrorDetails
// @Failure		500		{object}	ports.ErrorDetails
// @Router			/api/v1/{{.KebabCase}}/ [post]
func (ctrl *{{.CamelCase}}Controller) Create{{.PascalCase}}(c *ports.HttpContext) error {
	span, ctx := apm.StartSpan(c.Context(), "Create{{.PascalCase}}", "controller")
	defer span.End()

	var inputRequest create{{.PascalCase}}Request
	if err := c.BodyParser(&inputRequest); err != nil {
		ctrl.log.Error(ctx, err.Error())
		return appErr.NewBadRequestErr(err)
	}

	validationErrs := ctrl.customValidation.ValidateStruct(inputRequest)
	if validationErrs != nil {
		return utils.WithFieldErrors(
			appErr.NewBadRequestErr(errors.New(consts.ErrValidationFailedMsg)),
			validationErrs...,
		)
	}
	
	// TODO: Map inputRequest to a dto.{{.PascalCase}} struct.
	// Example:
	// entityDto := dto.{{.PascalCase}}{
	// 	Name: inputRequest.Name,
	// }
	var entityDto dto.{{.PascalCase}}


	createdEntity, err := ctrl.{{.CamelCase}}Service.Create{{.PascalCase}}(ctx, entityDto)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(ports.Response{
		Status: true,
		Data:   createdEntity,
	})
}

// @Summary		Get {{.PascalCase}} by ID
// @Description	This route will fetch a specific {{.LowerCase}} by its ID
// @Tags			{{.PascalCase}}
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"{{.PascalCase}} ID"
// @Success		200	{object}	ports.Response{data=dto.{{.PascalCase}}}
// @Failure		400	{object}	ports.ErrorDetails
// @Failure		404	{object}	ports.ErrorDetails
// @Failure		500	{object}	ports.ErrorDetails
// @Router			/api/v1/{{.KebabCase}}/{id} [get]
func (ctrl *{{.CamelCase}}Controller) Get{{.PascalCase}}ByID(c *ports.HttpContext) error {
	span, ctx := apm.StartSpan(c.Context(), "Get{{.PascalCase}}ByID", "controller")
	defer span.End()

	id, err := c.ParamsInt("id")
	if err != nil {
		return appErr.NewBadRequestErr(err)
	}

	entity, err := ctrl.{{.CamelCase}}Service.Get{{.PascalCase}}ByID(ctx, int64(id))
	if err != nil {
		return err
	}

	return c.JSON(ports.Response{
		Status: true,
		Data:   entity,
	})
}

// @Summary		Update a {{.PascalCase}}
// @Description	This route will update a {{.LowerCase}}
// @Tags			{{.PascalCase}}
// @Accept			json
// @Produce		json
// @Param			id		path		int	true	"{{.PascalCase}} ID"
// @Param			body	body		update{{.PascalCase}}Request 	true	"Update {{.PascalCase}} request"
// @Success		200		{object}	ports.Response{data=dto.{{.PascalCase}}}
// @Failure		400		{object}	ports.ErrorDetails
// @Failure		422		{object}	ports.ErrorDetails
// @Failure		500		{object}	ports.ErrorDetails
// @Router			/api/v1/{{.KebabCase}}/{id} [put]
func (ctrl *{{.CamelCase}}Controller) Update{{.PascalCase}}(c *ports.HttpContext) error {
	span, ctx := apm.StartSpan(c.Context(), "Update{{.PascalCase}}", "controller")
	defer span.End()

	id, err := c.ParamsInt("id")
	if err != nil {
		return appErr.NewBadRequestErr(err)
	}

	var inputRequest update{{.PascalCase}}Request
	if err := c.BodyParser(&inputRequest); err != nil {
		ctrl.log.Error(ctx, err.Error())
		return appErr.NewBadRequestErr(err)
	}

	validationErrs := ctrl.customValidation.ValidateStruct(inputRequest)
	if validationErrs != nil {
		return utils.WithFieldErrors(
			appErr.NewBadRequestErr(errors.New(consts.ErrValidationFailedMsg)),
			validationErrs...,
		)
	}
	
	// TODO: Map inputRequest to a dto.{{.PascalCase}} struct.
	// Example:
	// entityDto := dto.{{.PascalCase}}{
	// 	Name: inputRequest.Name,
	// }
	var entityDto dto.{{.PascalCase}}
	entityDto.ID = int64(id) // Set ID from path

	result, err := ctrl.{{.CamelCase}}Service.Update{{.PascalCase}}(ctx, entityDto)
	if err != nil {
		return err
	}

	return c.JSON(ports.Response{
		Status: true,
		Data:   result,
	})
}

// @Summary		Delete a {{.PascalCase}}
// @Description	This route will delete a {{.LowerCase}}
// @Tags			{{.PascalCase}}
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"{{.PascalCase}} ID"
// @Success		204
// @Failure		400	{object}	ports.ErrorDetails
// @Failure		500	{object}	ports.ErrorDetails
// @Router			/api/v1/{{.KebabCase}}/{id} [delete]
func (ctrl *{{.CamelCase}}Controller) Delete{{.PascalCase}}(c *ports.HttpContext) error {
	span, ctx := apm.StartSpan(c.Context(), "Delete{{.PascalCase}}", "controller")
	defer span.End()

	id, err := c.ParamsInt("id")
	if err != nil {
		return appErr.NewBadRequestErr(err)
	}

	err = ctrl.{{.CamelCase}}Service.Delete{{.PascalCase}}(ctx, int64(id))
	if err != nil {
		return err
	}

	return c.SendStatus(204)
}

// @Summary		Get All {{.PascalCase}}s
// @Description	Get all paginated {{.LowerCase}}s
// @Tags			{{.PascalCase}}
// @Accept			json
// @Produce		json
// @Param			params	query		httpUtils.ListRequest	false	"Pagination and filter parameters"
// @Success		200		{object}	ports.Response{data=[]dto.{{.PascalCase}}}
// @Failure		400	{object}	ports.ErrorDetails
// @Failure		500	{object}	ports.ErrorDetails
// @Router			/api/v1/{{.KebabCase}}/ [get]
func (ctrl *{{.CamelCase}}Controller) GetPaginated{{.PascalCase}}s(c *ports.HttpContext) error {
	span, ctx := apm.StartSpan(c.Context(), "GetPaginated{{.PascalCase}}s", "controller")
	defer span.End()
	
	// IMPORTANT: Define your filterable and sortable columns here
	columnMapping := map[string]string{
		// "fieldNameInQuery": "db_column_name",
		// "name": "title",
	}

	pagination, err := httpUtils.ParseAndValidatePagination(ctx, c, ctrl.customValidation, ctrl.log, columnMapping)
	if err != nil {
		return err
	}

	paginatedResult, resultPagination, err := ctrl.{{.CamelCase}}Service.GetPaginated{{.PascalCase}}s(ctx, pagination)
	if err != nil {
		return err
	}

	resp := ports.Response{
		Data: paginatedResult,
		Meta: &ports.Meta{
			Pagination: &resultPagination.Pagination,
		},
	}

	return c.JSON(resp)
}
`
