package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pulsar-beam/src/icrypto"
	. "github.com/pulsar-beam/src/middleware"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

func mockWorkerHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	w.WriteHeader(http.StatusOK)
	return
}

func TestAuthJWTMiddleware(t *testing.T) {
	// thanks goodness it is singleton
	publicKeyPath := "./example_public_key.pub"
	privateKeyPath := "./example_private_key"
	icrypto.NewRSAKeyPair(privateKeyPath, publicKeyPath)

	handlerTest := AuthVerifyJWT(http.HandlerFunc(mockHandler))

	req, err := http.NewRequest(http.MethodGet, "http://test", nil)
	errNil(t, err)

	rr := httptest.NewRecorder()

	// test missing authorization header
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusUnauthorized, rr.Code)

	// test a valid token
	req.Header.Set("Authorization", "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA")
	rr = httptest.NewRecorder()
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

	// test invalid token
	req.Header.Set("Authorization", "eeyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA")
	rr = httptest.NewRecorder()
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusUnauthorized, rr.Code)

	// test valid bearer token
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA")
	rr = httptest.NewRecorder()
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)
}

func TestAuthHeaderRequiredMiddleware(t *testing.T) {
	handlerTest := AuthHeaderRequired(http.HandlerFunc(mockHandler))

	req, err := http.NewRequest(http.MethodGet, "http://test", nil)
	errNil(t, err)

	rr := httptest.NewRecorder()

	// test missing authorization header
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusUnauthorized, rr.Code)

	// test a valid token
	req.Header.Set("Authorization", "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA")
	rr = httptest.NewRecorder()
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

	// test invalid token
	req.Header.Set("Authorization", "eeyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA")
	rr = httptest.NewRecorder()
	handlerTest.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)
}

func TestRateLimitMiddleware(t *testing.T) {
	handlerTest := LimitRate(http.HandlerFunc(mockHandler))

	req, err := http.NewRequest(http.MethodGet, "http://test", nil)
	errNil(t, err)

	request := func() {
		rr := httptest.NewRecorder()
		handlerTest.ServeHTTP(rr, req)
		equals(t, http.StatusOK, rr.Code)
	}
	//
	for i := 0; i < 1000; i++ {
		go request()
	}

}

func TestSemaphore(t *testing.T) {
	var sema = NewSema(2)
	err := sema.Release()
	equals(t, "all semaphore buffer empty", err.Error())

	err = sema.Acquire()
	errNil(t, err)
	sema.Acquire()
	err = sema.Acquire()
	assertErr(t, "all semaphore buffer full", err)

	sema.Release()
	errNil(t, sema.Acquire())

	errNil(t, sema.Release())
	errNil(t, sema.Release())

	err = sema.Release()
	assertErr(t, "all semaphore buffer empty", err)
}
