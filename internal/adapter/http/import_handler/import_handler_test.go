package http

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeImportUseCase struct {
	called bool
	err    error
}

func (f *fakeImportUseCase) Execute(
	ctx context.Context,
	r io.Reader,
) error {
	f.called = true
	return f.err
}

func TestImportHandler_Success(t *testing.T) {
	fake := &fakeImportUseCase{}
	handler := NewImportHandler(fake)

	var body strings.Builder
	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("file", "transacoes.csv")
	if err != nil {
		t.Fatal(err)
	}

	csv := "date,description,amount,type\n" +
		"2026-10-01,Salario,5000.00,INCOME\n"

	if _, err := io.Copy(fileWriter, strings.NewReader(csv)); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/transactions/import",
		strings.NewReader(body.String()),
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !fake.called {
		t.Fatal("expected use case to be called")
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}
}

func TestImportHandler_UseCaseError(t *testing.T) {
	fake := &fakeImportUseCase{
		err: errors.New("database failure"),
	}
	handler := NewImportHandler(fake)

	var body strings.Builder
	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("file", "transacoes.csv")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(
		fileWriter,
		"date,description,amount,type\n",
	); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/transactions/import",
		strings.NewReader(body.String()),
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !fake.called {
		t.Fatal("expected use case to be called")
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
