package integration

import (
	"context"
	"testing"
	"time"

	appdomain "cdk-office/internal/app/domain"
	docdomain "cdk-office/internal/document/domain"
	empdomain "cdk-office/internal/employee/domain"
	"cdk-office/internal/app/service"
	docservice "cdk-office/internal/document/service"
	empservice "cdk-office/internal/employee/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

// TestAppWithQRCodeIntegration tests the integration between AppService and QRCodeService
func TestAppWithQRCodeIntegration(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&appdomain.QRCode{}, &appdomain.Application{})

	appService := service.NewAppServiceWithDB(db)
	qrCodeService := service.NewQRCodeService()

	ctx := context.Background()

	// Create an application
	appReq := &service.CreateApplicationRequest{
		TeamID:      "team1",
		Name:        "QR Code App",
		Description: "Application for QR codes",
		Type:        "qrcode",
		Config:      "{}",
		CreatedBy:   "user1",
	}
	createdApp, err := appService.CreateApplication(ctx, appReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdApp)

	// Create a QR code associated with the application
	qrReq := &service.CreateQRCodeRequest{
		AppID:     createdApp.ID,
		Name:      "Test QR Code",
		Content:   "https://example.com",
		Type:      "static",
		URL:       "https://example.com",
		CreatedBy: "user1",
	}
	createdQR, err := qrCodeService.CreateQRCode(ctx, qrReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdQR)
	assert.Equal(t, createdApp.ID, createdQR.AppID)

	// Update the application
	updateAppReq := &service.UpdateApplicationRequest{
		Name:        "Updated QR Code App",
		Description: "Updated application for QR codes",
	}
	err = appService.UpdateApplication(ctx, createdApp.ID, updateAppReq)
	assert.NoError(t, err)

	// Verify the application was updated
	updatedApp, err := appService.GetApplication(ctx, createdApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated QR Code App", updatedApp.Name)
	assert.Equal(t, "Updated application for QR codes", updatedApp.Description)

	// Update the QR code
	updateQRReq := &service.UpdateQRCodeRequest{
		Name:    "Updated QR Code",
		Content: "https://updated-example.com",
		URL:     "https://updated-example.com",
	}
	err = qrCodeService.UpdateQRCode(ctx, createdQR.ID, updateQRReq)
	assert.NoError(t, err)

	// Verify the QR code was updated
	updatedQR, err := qrCodeService.GetQRCode(ctx, createdQR.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated QR Code", updatedQR.Name)
	assert.Equal(t, "https://updated-example.com", updatedQR.Content)
	assert.Equal(t, "https://updated-example.com", updatedQR.URL)

	// List QR codes for the application
	qrCodes, total, err := qrCodeService.ListQRCodes(ctx, createdApp.ID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, qrCodes, 1)
	assert.Equal(t, createdQR.ID, qrCodes[0].ID)

	// Delete the QR code
	err = qrCodeService.DeleteQRCode(ctx, createdQR.ID)
	assert.NoError(t, err)

	// Verify the QR code was deleted
	_, err = qrCodeService.GetQRCode(ctx, createdQR.ID)
	assert.Error(t, err)
	assert.Equal(t, "QR code not found", err.Error())

	// Delete the application
	err = appService.DeleteApplication(ctx, createdApp.ID)
	assert.NoError(t, err)

	// Verify the application was deleted
	_, err = appService.GetApplication(ctx, createdApp.ID)
	assert.Error(t, err)
	assert.Equal(t, "application not found", err.Error())
}

// TestAppWithFormIntegration tests the integration between AppService and FormService
func TestAppWithFormIntegration(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&appdomain.FormData{}, &appdomain.Application{})

	appService := service.NewAppServiceWithDB(db)
	formService := service.NewFormService()

	ctx := context.Background()

	// Create an application
	appReq := &service.CreateApplicationRequest{
		TeamID:      "team1",
		Name:        "Form App",
		Description: "Application for forms",
		Type:        "form",
		Config:      "{}",
		CreatedBy:   "user1",
	}
	createdApp, err := appService.CreateApplication(ctx, appReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdApp)

	// Create a form associated with the application
	formReq := &service.CreateFormRequest{
		AppID:       createdApp.ID,
		Name:        "Test Form",
		Description: "Test form description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		CreatedBy:   "user1",
	}
	createdForm, err := formService.CreateForm(ctx, formReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdForm)
	assert.Equal(t, createdApp.ID, createdForm.AppID)

	// Update the application
	updateAppReq := &service.UpdateApplicationRequest{
		Name:        "Updated Form App",
		Description: "Updated application for forms",
	}
	err = appService.UpdateApplication(ctx, createdApp.ID, updateAppReq)
	assert.NoError(t, err)

	// Verify the application was updated
	updatedApp, err := appService.GetApplication(ctx, createdApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Form App", updatedApp.Name)
	assert.Equal(t, "Updated application for forms", updatedApp.Description)

	// Update the form
	updateFormReq := &service.UpdateFormRequest{
		Name:        "Updated Form",
		Description: "Updated form description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}}`,
	}
	err = formService.UpdateForm(ctx, createdForm.ID, updateFormReq)
	assert.NoError(t, err)

	// Verify the form was updated
	updatedForm, err := formService.GetForm(ctx, createdForm.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Form", updatedForm.Name)
	assert.Equal(t, "Updated form description", updatedForm.Description)
	assert.Equal(t, `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}}`, updatedForm.Schema)

	// List forms for the application
	forms, total, err := formService.ListForms(ctx, createdApp.ID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, forms, 1)
	assert.Equal(t, createdForm.ID, forms[0].ID)

	// Submit form data
	submitReq := &service.SubmitFormDataRequest{
		FormID:    createdForm.ID,
		Data:      `{"name": "John Doe", "age": 30}`,
		CreatedBy: "user1",
	}
	formEntry, err := formService.SubmitFormData(ctx, submitReq)
	assert.NoError(t, err)
	assert.NotNil(t, formEntry)

	// List form data entries
	entries, total, err := formService.ListFormDataEntries(ctx, createdForm.ID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, entries, 1)
	assert.Equal(t, formEntry.ID, entries[0].ID)

	// Delete the form
	err = formService.DeleteForm(ctx, createdForm.ID)
	assert.NoError(t, err)

	// Verify the form was deleted
	_, err = formService.GetForm(ctx, createdForm.ID)
	assert.Error(t, err)
	assert.Equal(t, "form not found", err.Error())

	// Delete the application
	err = appService.DeleteApplication(ctx, createdApp.ID)
	assert.NoError(t, err)

	// Verify the application was deleted
	_, err = appService.GetApplication(ctx, createdApp.ID)
	assert.Error(t, err)
	assert.Equal(t, "application not found", err.Error())
}

// TestDocumentWithEmployeeIntegration tests the integration between DocumentService and EmployeeService
func TestDocumentWithEmployeeIntegration(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&docdomain.Document{}, &docdomain.DocumentVersion{}, &empdomain.Employee{}, &empdomain.Department{})

	docService := docservice.NewDocumentServiceWithDB(db)
	empService := empservice.NewEmployeeServiceWithDB(db)

	ctx := context.Background()

	// Create a department
	dept := &empdomain.Department{
		ID:   "dept-001",
		Name: "技术部",
	}
	err := db.Create(dept).Error
	assert.NoError(t, err)

	// Create an employee
	empReq := &empservice.CreateEmployeeRequest{
		UserID:     "user-001",
		TeamID:     "team-001",
		DeptID:     "dept-001",
		EmployeeID: "emp-001",
		RealName:   "张三",
		Gender:     "男",
		BirthDate:  time.Now().AddDate(-30, 0, 0),
		HireDate:   time.Now(),
		Position:   "软件工程师",
	}
	createdEmployee, err := empService.CreateEmployee(ctx, empReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdEmployee)

	// Create a document owned by the employee
	docReq := &docservice.UploadRequest{
		Title:       "Employee Document",
		Description: "Document owned by employee",
		FilePath:    "/path/to/document.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     createdEmployee.UserID,
		TeamID:      createdEmployee.TeamID,
		Tags:        "employee,document",
	}
	createdDoc, err := docService.Upload(ctx, docReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdDoc)
	assert.Equal(t, createdEmployee.UserID, createdDoc.OwnerID)
	assert.Equal(t, createdEmployee.TeamID, createdDoc.TeamID)

	// Update the employee
	updateEmpReq := &empservice.UpdateEmployeeRequest{
		Position: "高级软件工程师",
	}
	err = empService.UpdateEmployee(ctx, createdEmployee.ID, updateEmpReq)
	assert.NoError(t, err)

	// Verify the employee was updated
	updatedEmployee, err := empService.GetEmployee(ctx, createdEmployee.ID)
	assert.NoError(t, err)
	assert.Equal(t, "高级软件工程师", updatedEmployee.Position)

	// Update the document
	updateDocReq := &docservice.UpdateRequest{
		Title:       "Updated Employee Document",
		Description: "Updated document owned by employee",
		Status:      "archived",
	}
	err = docService.UpdateDocument(ctx, createdDoc.ID, updateDocReq)
	assert.NoError(t, err)

	// Verify the document was updated
	updatedDoc, err := docService.GetDocument(ctx, createdDoc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Employee Document", updatedDoc.Title)
	assert.Equal(t, "Updated document owned by employee", updatedDoc.Description)
	assert.Equal(t, "archived", updatedDoc.Status)

	// List documents for the employee's team using SearchService
	searchService := docservice.NewSearchService()
	docs, total, err := searchService.SearchDocuments(ctx, "", createdEmployee.TeamID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, docs, 1)
	assert.Equal(t, createdDoc.ID, docs[0].ID)

	// Delete the document
	err = docService.DeleteDocument(ctx, createdDoc.ID)
	assert.NoError(t, err)

	// Verify the document was deleted
	_, err = docService.GetDocument(ctx, createdDoc.ID)
	assert.Error(t, err)
	assert.Equal(t, "document not found", err.Error())

	// Delete the employee
	err = empService.DeleteEmployee(ctx, createdEmployee.ID)
	assert.NoError(t, err)

	// Verify the employee was deleted
	_, err = empService.GetEmployee(ctx, createdEmployee.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "employee not found")
}