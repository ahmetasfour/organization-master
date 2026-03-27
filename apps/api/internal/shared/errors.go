package shared

import "errors"

// Sentinel errors for common error conditions
var (
	// ErrApplicationTerminated indicates an application has been terminated and cannot be modified
	ErrApplicationTerminated = errors.New("başvuru sonlandırılmıştır ve değiştirilemez")

	// ErrTokenNotFound indicates the token was not found in the database
	ErrTokenNotFound = errors.New("token bulunamadı")

	// ErrTokenExpired indicates a token has expired
	ErrTokenExpired = errors.New("token süresi dolmuştur")

	// ErrTokenUsed indicates a token has already been used
	ErrTokenUsed = errors.New("token daha önce kullanılmıştır")

	// ErrInvalidCredentials indicates invalid login credentials
	ErrInvalidCredentials = errors.New("geçersiz email veya şifre")

	// ErrForbidden indicates the user lacks permission for the requested action
	ErrForbidden = errors.New("bu işlem için yetkiniz bulunmamaktadır")

	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("kayıt bulunamadı")

	// ErrDuplicateVote indicates a user has already voted in a stage
	ErrDuplicateVote = errors.New("bu aşamada zaten oy kullandınız")

	// ErrInvalidTransition indicates an invalid state transition
	ErrInvalidTransition = errors.New("geçersiz durum değişikliği")

	// ErrReferenceQuotaExceeded indicates too many references have been requested
	ErrReferenceQuotaExceeded = errors.New("referans kotası aşıldı")

	// ErrUnauthorized indicates the request lacks valid authentication
	ErrUnauthorized = errors.New("yetkilendirme başarısız")
)
