package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/plutoframework/pluto-auth/methods/password"
)

func TestDevelopmentIngressUsesSocketPeerAndBoundsRate(t *testing.T) {
	ingress, err := DevelopmentIngress()
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	handler := ingress(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) { calls++; w.WriteHeader(204) },
		),
	)

	for _, peer := range []string{"203.0.113.9:1234", "invalid", "localhost:1234"} {
		r := httptest.NewRequest("POST", "/auth/signin", nil)
		r.RemoteAddr = peer
		r.Header.Set("X-Forwarded-For", "127.0.0.1")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != 403 || calls != 0 {
			t.Fatal("forwarded input established client trust")
		}
	}

	for i := 0; i < 121; i++ {
		r := httptest.NewRequest("POST", "/auth/signin", nil)
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("X-Forwarded-For", string(rune(i+65)))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		want := 204
		if i == 120 {
			want = 429
		}

		if w.Code != want {
			t.Fatal("per-client limit bypassed")
		}
	}
}

func TestDevelopmentIngressConcurrencyAndCancellation(t *testing.T) {
	ingress, err := DevelopmentIngress()
	if err != nil {
		t.Fatal(err)
	}

	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	handler := ingress(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := r.Context().Deadline(); !ok {
				t.Error("request timeout missing")
			}

			entered <- struct{}{}

			<-release
			w.WriteHeader(204)
		}),
	)

	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Add(1)

		go func() {
			defer group.Done()

			r := httptest.NewRequest("POST", "/auth/signin", nil)
			r.RemoteAddr = "127.0.0.1:1234"
			handler.ServeHTTP(httptest.NewRecorder(), r)
		}()
	}

	for i := 0; i < 4; i++ {
		<-entered
	}

	r := httptest.NewRequest("POST", "/auth/signin", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	close(release)
	group.Wait()

	if w.Code != 429 {
		t.Fatal("concurrency overflow queued or admitted")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r.WithContext(ctx))

	if w.Code != 429 {
		t.Fatal("cancelled request admitted")
	}
}

func FuzzDevelopmentIngressPeer(f *testing.F) {
	for _, peer := range []string{"127.0.0.1:1234", "[::1]:1234", "203.0.113.9:1234", "invalid"} {
		f.Add(peer)
	}

	f.Fuzz(func(t *testing.T, peer string) {
		ingress, err := DevelopmentIngress()
		if err != nil {
			t.Fatal(err)
		}

		called := false
		h := ingress(
			http.HandlerFunc(
				func(w http.ResponseWriter, _ *http.Request) { called = true; w.WriteHeader(204) },
			),
		)
		r := httptest.NewRequest("POST", "/auth/signin", nil)
		r.RemoteAddr = peer
		r.Header.Set("X-Forwarded-For", "127.0.0.1")

		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)

		if (w.Code == 204) != called || (w.Code != 204 && w.Code != 403) {
			t.Fatal("ingress result mismatch")
		}
	})
}

func TestDevelopmentSenderRejectsExternalDelivery(t *testing.T) {
	for _, address := range []string{"smtp.example.test:25", "203.0.113.1:25", "localhost:1025", "127.0.0.1:0"} {
		if _, err := NewDevelopmentSMTP(address); err == nil {
			t.Fatal("non-loopback or invalid SMTP accepted")
		}
	}
}

func TestDevelopmentSenderCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	sender, err := NewDevelopmentSMTP(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	accepted := make(chan net.Conn, 1)

	go func() {
		connection, err := listener.Accept()
		if err == nil {
			accepted <- connection
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- sender.Send(ctx, password.Notification{
			Email: "synthetic@example.test", Purpose: password.VerifyEmail,
			Token: base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
		})
	}()

	select {
	case connection := <-accepted:
		defer connection.Close()
	case <-time.After(time.Second):
		t.Fatal("SMTP did not connect to local sink")
	}

	// The sink deliberately sends no SMTP greeting. Cancellation must close
	// the blocked socket rather than wait indefinitely for the server.
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, password.ErrDelivery) {
			t.Fatal("unredacted delivery error")
		}
	case <-time.After(time.Second):
		t.Fatal("delivery ignored cancellation")
	}
}
