package applications

import (
	"fmt"

	"membership-system/api/internal/shared"
)

// transitions defines all valid state transitions per membership type
var transitions = map[MembershipType]map[ApplicationStatus][]ApplicationStatus{

	// Asil & Akademik share the same flow
	MembershipAsil: {
		StatusBasvuruAlindi:      {StatusReferansBekleniyor},
		StatusReferansBekleniyor: {StatusReferansTamamlandi, StatusReferansRed},
		StatusReferansTamamlandi: {StatusYKOnIncelemede},
		StatusYKOnIncelemede:     {StatusOnOnaylandi, StatusYKRed},
		StatusOnOnaylandi:        {StatusItibarTaramasinda},
		StatusItibarTaramasinda:  {StatusItibarTemiz, StatusItibarRed},
		StatusItibarTemiz:        {StatusGundemde},
		StatusGundemde:           {StatusKabul, StatusReddedildi},
		// Terminal states — no further transitions
		StatusReferansRed: {},
		StatusYKRed:       {},
		StatusItibarRed:   {},
		StatusReddedildi:  {},
		StatusKabul:       {},
	},

	MembershipAkademik: {
		StatusBasvuruAlindi:      {StatusReferansBekleniyor},
		StatusReferansBekleniyor: {StatusReferansTamamlandi, StatusReferansRed},
		StatusReferansTamamlandi: {StatusYKOnIncelemede},
		StatusYKOnIncelemede:     {StatusOnOnaylandi, StatusYKRed},
		StatusOnOnaylandi:        {StatusItibarTaramasinda},
		StatusItibarTaramasinda:  {StatusItibarTemiz, StatusItibarRed},
		StatusItibarTemiz:        {StatusGundemde},
		StatusGundemde:           {StatusKabul, StatusReddedildi},
		StatusReferansRed:        {},
		StatusYKRed:              {},
		StatusItibarRed:          {},
		StatusReddedildi:         {},
		StatusKabul:              {},
	},

	// Profesyonel & Öğrenci flow
	MembershipProfesyonel: {
		StatusBasvuruAlindi:    {StatusDanismaSurecinde},
		StatusDanismaSurecinde: {StatusGundemde, StatusDanismaRed},
		StatusGundemde:         {StatusKabul, StatusReddedildi},
		StatusDanismaRed:       {},
		StatusReddedildi:       {},
		StatusKabul:            {},
	},

	MembershipOgrenci: {
		StatusBasvuruAlindi:    {StatusDanismaSurecinde},
		StatusDanismaSurecinde: {StatusGundemde, StatusDanismaRed},
		StatusGundemde:         {StatusKabul, StatusReddedildi},
		StatusDanismaRed:       {},
		StatusReddedildi:       {},
		StatusKabul:            {},
	},

	// Onursal flow
	MembershipOnursal: {
		StatusOneriAlindi:        {StatusYKOnIncelemede},
		StatusYKOnIncelemede:     {StatusOnOnaylandi, StatusYKRed},
		StatusOnOnaylandi:        {StatusYIKDegerlendirmede},
		StatusYIKDegerlendirmede: {StatusGundemde, StatusYIKRed},
		StatusGundemde:           {StatusKabul, StatusReddedildi},
		StatusYKRed:              {},
		StatusYIKRed:             {},
		StatusReddedildi:         {},
		StatusKabul:              {},
	},
}

// terminalStatuses holds all statuses from which no further transitions exist
var terminalStatuses = map[ApplicationStatus]bool{
	StatusReferansRed: true,
	StatusYKRed:       true,
	StatusItibarRed:   true,
	StatusDanismaRed:  true,
	StatusYIKRed:      true,
	StatusReddedildi:  true,
	StatusKabul:       true,
}

// initialStatuses maps each membership type to its starting status
var initialStatuses = map[MembershipType]ApplicationStatus{
	MembershipAsil:        StatusBasvuruAlindi,
	MembershipAkademik:    StatusBasvuruAlindi,
	MembershipProfesyonel: StatusBasvuruAlindi,
	MembershipOgrenci:     StatusBasvuruAlindi,
	MembershipOnursal:     StatusOneriAlindi,
}

// ValidateTransition checks whether a transition from currentStatus to nextStatus
// is allowed for the given membership type. Returns ErrInvalidTransition on failure.
func ValidateTransition(membershipType MembershipType, currentStatus, nextStatus ApplicationStatus) error {
	allowed, ok := transitions[membershipType]
	if !ok {
		return fmt.Errorf("%w: unknown membership type %q", shared.ErrInvalidTransition, membershipType)
	}

	targets, ok := allowed[currentStatus]
	if !ok {
		return fmt.Errorf("%w: status %q has no defined transitions for type %q",
			shared.ErrInvalidTransition, currentStatus, membershipType)
	}

	for _, t := range targets {
		if t == nextStatus {
			return nil
		}
	}

	return fmt.Errorf("%w: cannot move from %q to %q for membership type %q",
		shared.ErrInvalidTransition, currentStatus, nextStatus, membershipType)
}

// IsTerminated returns true if the status is a terminal (dead-end) state.
func IsTerminated(status ApplicationStatus) bool {
	return terminalStatuses[status]
}

// GetInitialStatus returns the first status for a given membership type.
func GetInitialStatus(membershipType MembershipType) ApplicationStatus {
	if s, ok := initialStatuses[membershipType]; ok {
		return s
	}
	return StatusBasvuruAlindi
}

// AllowedTransitions returns the list of statuses that can follow currentStatus.
// Returns nil if the membership type is unknown or if in a terminal state.
func AllowedTransitions(membershipType MembershipType, currentStatus ApplicationStatus) []ApplicationStatus {
	allowed, ok := transitions[membershipType]
	if !ok {
		return nil
	}
	return allowed[currentStatus]
}
