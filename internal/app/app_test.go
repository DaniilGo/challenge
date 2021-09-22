package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DaniilGo/challenge/internal/merge"
	"github.com/DaniilGo/challenge/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testURL  = "someurl"
	testBody = "somebody"
)

func TestRun(t *testing.T) {
	rtMock := &roundTripperMock{}
	rtMock.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		assert.Equal(t, r.Method, http.MethodPost)
		assert.Equal(t, r.URL.String(), testURL)
		assert.Equal(t, r.Header, http.Header{"Content-Type": []string{"application/json"}})
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, testBody, string(b))
		return true
	})).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"somekey":"somevalue"}`)),
	}, nil).Once()

	httpClient := &http.Client{
		Transport: rtMock,
	}

	s := source.NewHttpPostSource(testURL, []byte(testBody), httpClient)
	m := merge.NewJSONSprintfMerger(`{"outerkey": %s}`)
	b := &bytes.Buffer{}
	a := NewApp(s, m, b)

	err := a.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "{\n \"outerkey\": {\n  \"somekey\": \"somevalue\"\n }\n}", b.String())
}

func TestRunFailed_GetReadCloser(t *testing.T) {
	wantErr := errors.New("some error")

	rtMock := &roundTripperMock{}
	rtMock.On("RoundTrip", mock.Anything).Return((*http.Response)(nil), wantErr).Once()

	httpClient := &http.Client{
		Transport: rtMock,
	}

	s := source.NewHttpPostSource(testURL, []byte(testBody), httpClient)
	a := NewApp(s, nil, nil)

	err := a.Run(context.Background())
	assert.Equal(t, "failed to get read closer: httpPostSource.Do(req) error: Post \"someurl\": some error", err.Error())
}

func TestRunFailed_Read(t *testing.T) {
	wantErr := errors.New("some error")
	rtMock := &roundTripperMock{}
	rtMock.On("RoundTrip", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       &errReadCloser{readErr: wantErr},
	}, nil).Once()

	httpClient := &http.Client{
		Transport: rtMock,
	}

	s := source.NewHttpPostSource(testURL, []byte(testBody), httpClient)
	a := NewApp(s, nil, nil)

	err := a.Run(context.Background())
	assert.Equal(t, "failed to read all: some error", err.Error())
}

func TestRun_GetMerged(t *testing.T) {
	rtMock := &roundTripperMock{}
	rtMock.On("RoundTrip", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"somekey":"somevalue"}`)),
	}, nil).Once()

	httpClient := &http.Client{
		Transport: rtMock,
	}

	s := source.NewHttpPostSource(testURL, []byte(testBody), httpClient)
	m := merge.NewJSONSprintfMerger(`{"outerkey": %s}`)
	b := &bytes.Buffer{}
	a := NewApp(s, m, b)

	err := a.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "{\n \"outerkey\": {\n  \"somekey\": \"somevalue\"\n }\n}", b.String())
}

func TestRunFailed_GetMerged(t *testing.T) {
	rtMock := &roundTripperMock{}
	rtMock.On("RoundTrip", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`invalid json`)),
	}, nil).Once()

	httpClient := &http.Client{
		Transport: rtMock,
	}

	s := source.NewHttpPostSource(testURL, []byte(testBody), httpClient)
	m := merge.NewJSONSprintfMerger(`%s`)
	a := NewApp(s, m, nil)

	err := a.Run(context.Background())
	assert.Equal(t, "failed to get merged: invalid json", err.Error())
}

func TestRunFailed_Write(t *testing.T) {
	wantErr := errors.New("some error")
	rtMock := &roundTripperMock{}
	rtMock.On("RoundTrip", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"somekey":"somevalue"}`)),
	}, nil).Once()

	httpClient := &http.Client{
		Transport: rtMock,
	}

	s := source.NewHttpPostSource(testURL, []byte(testBody), httpClient)
	m := merge.NewJSONSprintfMerger(`%s`)
	a := NewApp(s, m, &errWriter{err: wantErr})

	err := a.Run(context.Background())
	assert.Equal(t, "failed to write: some error", err.Error())
}

func BenchmarkApp_Run(b *testing.B) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "0")
	}))
	ts.Start()
	defer ts.Close()

	s := source.NewHttpPostSource(ts.URL, []byte(""), ts.Client())
	m := merge.NewJSONSprintfMerger(`%s`)
	buf := &bytes.Buffer{}
	a := NewApp(s, m, buf)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		_ = a.Run(ctx)
	}
}

type roundTripperMock struct {
	mock.Mock
}

func (r *roundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	args := r.Called(req)
	if args.Get(0) == nil {
		return (*http.Response)(nil), args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

type errReadCloser struct {
	readErr  error
	closeErr error
}

func (e *errReadCloser) Read(_ []byte) (n int, err error) {
	if e.readErr != nil {
		return 0, e.readErr
	}
	return 1, nil
}

func (e *errReadCloser) Close() error {
	if e.closeErr != nil {
		return e.closeErr
	}
	return nil
}

type errWriter struct {
	err error
}

func (e *errWriter) Write(_ []byte) (n int, err error) {
	if e.err != nil {
		return 0, e.err
	}
	return 1, nil
}
