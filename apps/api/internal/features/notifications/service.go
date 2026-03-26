package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ─── Inline templates ──────────────────────────────────────────────────────────
// Templates are embedded here to keep the service self-contained.
// All templates are bilingual (Turkish) as required by spec.

const tmplReferenceRequest = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Referans Onayı Bekleniyor</h2>
  <p>Sayın <strong>{{.RefereeName}}</strong>,</p>
  <p>
    <strong>{{.ApplicantName}}</strong> adlı kişi
    <strong>{{.MembershipType}}</strong> üyeliği için başvurmuştur
    ve sizi referans olarak göstermiştir.
  </p>
  <p>Görüşünüzü bildirmek için aşağıdaki bağlantıya tıklayınız:</p>
  <p>
    <a href="{{.ResponseURL}}" style="
      display:inline-block;
      background:#1a73e8;
      color:#fff;
      padding:12px 24px;
      border-radius:4px;
      text-decoration:none;
      font-weight:600
    ">Görüşümü Bildir</a>
  </p>
  <p style="color:#666;font-size:0.875rem">
    Bu bağlantı <strong>{{.ExpiresAt}}</strong> tarihinde geçerliliğini yitirecektir.<br>
    Bağlantı yalnızca <strong>bir kez</strong> kullanılabilir.
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">
    Bu e-postayı yanlışlıkla aldıysanız lütfen dikkate almayınız.
  </p>
</body>
</html>`

const tmplNewRefNeeded = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Yeni Referans Gerekiyor</h2>
  <p>Sayın <strong>{{.ApplicantName}}</strong>,</p>
  <p>
    <strong>{{.UnknownRefereeName}}</strong> adlı referansınız sizi tanımadığını bildirmiştir.
  </p>
  <p>Başvurunuzun değerlendirmeye devam edebilmesi için lütfen sisteme giriş yaparak yeni bir referans ekleyiniz.</p>
  <p>
    <a href="{{.PortalURL}}" style="
      display:inline-block;
      background:#1a73e8;
      color:#fff;
      padding:12px 24px;
      border-radius:4px;
      text-decoration:none;
      font-weight:600
    ">Portala Git</a>
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">
    Üyelik Yönetim Sistemi — Bu e-postayı yanlışlıkla aldıysanız lütfen dikkate almayınız.
  </p>
</body>
</html>`

const tmplAccepted = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Başvurunuz Kabul Edildi</h2>
  <p>Sayın <strong>{{.ApplicantName}}</strong>,</p>
  <p>
    <strong>{{.MembershipType}}</strong> üyeliği için yaptığınız başvuru kabul edilmiştir.
    Üyeliğiniz hayırlı olsun.
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">Üyelik Yönetim Sistemi</p>
</body>
</html>`

const tmplConsultationRequest = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Üye Danışma Talebi</h2>
  <p>Sayın <strong>{{.MemberName}}</strong>,</p>
  <p>
    <strong>{{.ApplicantName}}</strong> adlı kişi
    <strong>{{.MembershipType}}</strong> üyeliği için başvurmuştur.
  </p>
  {{if .ApplicantLinkedIn}}
  <p>LinkedIn: <a href="{{.ApplicantLinkedIn}}">{{.ApplicantLinkedIn}}</a></p>
  {{end}}
  <p>Görüşünüzü bildirmek için aşağıdaki bağlantıya tıklayınız:</p>
  <p>
    <a href="{{.ResponseURL}}" style="
      display:inline-block;
      background:#1a73e8;
      color:#fff;
      padding:12px 24px;
      border-radius:4px;
      text-decoration:none;
      font-weight:600
    ">Görüşümü Bildir</a>
  </p>
  <p style="color:#666;font-size:0.875rem">
    Bu bağlantı <strong>{{.ExpiresAt}}</strong> tarihinde geçerliliğini yitirecektir.<br>
    Bağlantı yalnızca <strong>bir kez</strong> kullanılabilir.
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">
    Bu e-postayı yanlışlıkla aldıysanız lütfen dikkate almayınız.
  </p>
</body>
</html>`

// CRITICAL: rejection email must NOT include rejection_reason.
const tmplRejected = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Başvurunuz Hakkında Bilgilendirme</h2>
  <p>Sayın <strong>{{.ApplicantName}}</strong>,</p>
  <p>Üyelik başvurunuz değerlendirilmiş olup sonuçlanmıştır.</p>
  <p>Daha fazla bilgi için lütfen kurum ile iletişime geçiniz.</p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">Üyelik Yönetim Sistemi</p>
</body>
</html>`

const tmplReferenceReminder = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#c0392b">Hatırlatma: Referans Yanıtı Bekleniyor</h2>
  <p>Sayın <strong>{{.RefereeName}}</strong>,</p>
  <p>
    <strong>{{.ApplicantName}}</strong> adlı kişi için size iletilen referans talebini henüz yanıtlamadınız.
  </p>
  <p style="color:#c0392b;font-weight:600">
    Bu bağlantı yaklaşık <strong>{{.HoursRemaining}} saat</strong> içinde geçerliliğini yitirecektir.
  </p>
  <p>Görüşünüzü bildirmek için lütfen aşağıdaki bağlantıya tıklayınız:</p>
  <p>
    <a href="{{.ResponseURL}}" style="
      display:inline-block;
      background:#c0392b;
      color:#fff;
      padding:12px 24px;
      border-radius:4px;
      text-decoration:none;
      font-weight:600
    ">Şimdi Yanıtla</a>
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">
    Bu e-posta otomatik olarak gönderilmiştir. Üyelik Yönetim Sistemi
  </p>
</body>
</html>`

const tmplReputationQuery = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Üye Adayı Hakkında Bilgi Talebi</h2>
  <p>Sayın <strong>{{.ContactName}}</strong>,</p>
  <p>
    <strong>{{.ApplicantName}}</strong> adlı kişi hakkında üyelik başvurusu bulunmaktadır.
  </p>
  {{if .ApplicantLinkedIn}}
  <p>LinkedIn: <a href="{{.ApplicantLinkedIn}}">{{.ApplicantLinkedIn}}</a></p>
  {{end}}
  <p style="font-weight:600">SORU: Bu kişi hakkında olumsuz bir bilginiz var mı?</p>
  <p>Yanıtlamak için aşağıdaki bağlantıya tıklayınız:</p>
  <p>
    <a href="{{.ResponseURL}}" style="
      display:inline-block;
      background:#1a73e8;
      color:#fff;
      padding:12px 24px;
      border-radius:4px;
      text-decoration:none;
      font-weight:600
    ">Yanıtla</a>
  </p>
  <p style="color:#666;font-size:0.875rem">
    Bu bağlantı <strong>{{.ExpiresAt}}</strong> tarihinde geçerliliğini yitirecektir.<br>
    Bağlantı yalnızca <strong>bir kez</strong> kullanılabilir.
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">
    Bu e-posta otomatik olarak gönderilmiştir. Üyelik Yönetim Sistemi
  </p>
</body>
</html>`

const tmplHonoraryProposal = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;max-width:600px;margin:auto;padding:20px">
  <h2 style="color:#1a1a2e">Onursal Üyelik Önerisi</h2>
  <p>Sayın <strong>{{.YKMemberName}}</strong>,</p>
  <p>
    <strong>{{.ProposerName}}</strong> tarafından
    <strong>{{.NomineeName}}</strong> adlı kişi için onursal üyelik önerisi sunulmuştur.
  </p>
  {{if .NomineeLinkedIn}}
  <p>LinkedIn: <a href="{{.NomineeLinkedIn}}">{{.NomineeLinkedIn}}</a></p>
  {{end}}
  <p><strong>Öneri Gerekçesi:</strong></p>
  <blockquote style="border-left:4px solid #1a73e8;margin:0;padding:12px 16px;background:#f8f9fa;color:#333">
    {{.ProposalReason}}
  </blockquote>
  <br>
  <p>Başvuruyu incelemek için:</p>
  <p>
    <a href="{{.ReviewURL}}" style="
      display:inline-block;
      background:#1a73e8;
      color:#fff;
      padding:12px 24px;
      border-radius:4px;
      text-decoration:none;
      font-weight:600
    ">Başvuruyu İncele</a>
  </p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#999;font-size:0.75rem">
    Bu e-posta otomatik olarak gönderilmiştir. Üyelik Yönetim Sistemi
  </p>
</body>
</html>`

// ─── logRow (local minimal struct to avoid import cycles) ──────────────────────

type emailLogRow struct {
	ID         string         `gorm:"column:id"`
	Action     string         `gorm:"column:action"`
	ActorID    *string        `gorm:"column:actor_id"`
	ActorRole  string         `gorm:"column:actor_role"`
	EntityType string         `gorm:"column:entity_type"`
	EntityID   string         `gorm:"column:entity_id"`
	Metadata   datatypes.JSON `gorm:"column:metadata"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
}

func (emailLogRow) TableName() string { return "logs" }

// ─── Service ───────────────────────────────────────────────────────────────────

// Service sends notification emails and logs each dispatch.
type Service struct {
	mailer  *Mailer
	db      *gorm.DB
	baseURL string
}

// NewService creates a new notification service.
func NewService(mailer *Mailer, db *gorm.DB, baseURL string) *Service {
	return &Service{mailer: mailer, db: db, baseURL: baseURL}
}

// SendReferenceRequest sends a tokenized reference-request email to a referee.
func (s *Service) SendReferenceRequest(
	ctx context.Context,
	refID, refereeEmail, refereeName string,
	rawToken string,
	applicantName, membershipType string,
	expiresAt time.Time,
) error {
	responseURL := fmt.Sprintf("%s/respond/reference/%s", s.baseURL, rawToken)

	data := ReferenceRequestData{
		RefereeName:    refereeName,
		ApplicantName:  applicantName,
		MembershipType: membershipType,
		ResponseURL:    responseURL,
		ExpiresAt:      FormatTime(expiresAt),
	}

	html, text, err := Render(tmplReferenceRequest, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Referans Onayı Bekleniyor"
	if err := s.mailer.Send(refereeEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send reference request: %w", err)
	}

	s.logEmail(ctx, refID, "reference", "email.reference_request", refereeEmail)
	return nil
}

// SendNewRefNeeded notifies an applicant that one of their referees said "unknown".
func (s *Service) SendNewRefNeeded(
	ctx context.Context,
	appID, applicantEmail, applicantName, unknownRefereeName string,
) error {
	data := NewRefNeededData{
		ApplicantName:      applicantName,
		UnknownRefereeName: unknownRefereeName,
		PortalURL:          s.baseURL,
	}

	html, text, err := Render(tmplNewRefNeeded, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Yeni Referans Gerekiyor"
	if err := s.mailer.Send(applicantEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send new ref needed: %w", err)
	}

	s.logEmail(ctx, appID, "application", "email.new_ref_needed", applicantEmail)
	return nil
}

// SendConsultationRequest sends a tokenized consultation-request email to a member.
func (s *Service) SendConsultationRequest(
	ctx context.Context,
	memberEmail, memberName string,
	rawToken string,
	applicantName, membershipType, applicantLinkedIn string,
	expiresAt time.Time,
) error {
	responseURL := fmt.Sprintf("%s/respond/consultation/%s", s.baseURL, rawToken)

	data := ConsultationData{
		MemberName:        memberName,
		ApplicantName:     applicantName,
		MembershipType:    membershipType,
		ApplicantLinkedIn: applicantLinkedIn,
		ResponseURL:       responseURL,
		ExpiresAt:         FormatTime(expiresAt),
	}

	html, text, err := Render(tmplConsultationRequest, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Üye Danışma Talebi"
	if err := s.mailer.Send(memberEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send consultation request: %w", err)
	}

	s.logEmail(ctx, "", "consultation", "email.consultation_request", memberEmail)
	return nil
}

// SendAccepted notifies an applicant that their application was accepted.
func (s *Service) SendAccepted(
	ctx context.Context,
	appID, applicantEmail, applicantName, membershipType string,
) error {
	data := AcceptedData{ApplicantName: applicantName, MembershipType: membershipType}

	html, text, err := Render(tmplAccepted, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Başvurunuz Kabul Edildi"
	if err := s.mailer.Send(applicantEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send accepted: %w", err)
	}

	s.logEmail(ctx, appID, "application", "email.accepted", applicantEmail)
	return nil
}

// SendRejected notifies an applicant that their application was rejected.
// CRITICAL: rejection_reason is never included in this email.
func (s *Service) SendRejected(
	ctx context.Context,
	appID, applicantEmail, applicantName string,
) error {
	data := RejectedData{ApplicantName: applicantName}

	html, text, err := Render(tmplRejected, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Başvurunuz Hakkında Bilgilendirme"
	if err := s.mailer.Send(applicantEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send rejected: %w", err)
	}

	s.logEmail(ctx, appID, "application", "email.rejected", applicantEmail)
	return nil
}

// SendReferenceReminder sends a reminder to a referee whose token is near expiry.
func (s *Service) SendReferenceReminder(
	ctx context.Context,
	refID, refereeEmail, refereeName string,
	rawToken string,
	applicantName string,
	expiresAt time.Time,
) error {
	responseURL := fmt.Sprintf("%s/respond/reference/%s", s.baseURL, rawToken)
	hoursRemaining := int(time.Until(expiresAt).Hours())
	if hoursRemaining < 1 {
		hoursRemaining = 1
	}

	data := ReferenceReminderData{
		RefereeName:    refereeName,
		ApplicantName:  applicantName,
		ResponseURL:    responseURL,
		HoursRemaining: hoursRemaining,
	}

	html, text, err := Render(tmplReferenceReminder, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Hatırlatma: Referans Yanıtı Bekleniyor"
	if err := s.mailer.Send(refereeEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send reference reminder: %w", err)
	}

	s.logEmail(ctx, refID, "reference", "email.reminder_sent", refereeEmail)
	return nil
}

// SendReputationQuery sends a tokenized reputation-query email to an external contact.
func (s *Service) SendReputationQuery(
	ctx context.Context,
	contactID, contactEmail, contactName string,
	rawToken string,
	applicantName, applicantLinkedIn string,
	expiresAt time.Time,
) error {
	responseURL := fmt.Sprintf("%s/respond/reputation/%s", s.baseURL, rawToken)

	data := ReputationQueryData{
		ContactName:   contactName,
		ApplicantName: applicantName,
		ApplicantURL:  applicantLinkedIn,
		ResponseURL:   responseURL,
		ExpiresAt:     FormatTime(expiresAt),
	}

	html, text, err := Render(tmplReputationQuery, data)
	if err != nil {
		return err
	}

	subject := "[Membership System] Üye Adayı Hakkında Bilgi Talebi"
	if err := s.mailer.Send(contactEmail, subject, html, text); err != nil {
		return fmt.Errorf("notifications: send reputation query: %w", err)
	}

	s.logEmail(ctx, contactID, "reputation_contact", "email.reputation_query", contactEmail)
	return nil
}

// SendHonoraryProposal notifies all YK members when an honorary membership proposal is submitted.
// ykMembers is a list of (id, email, full_name) tuples — one email per YK member.
func (s *Service) SendHonoraryProposal(
	ctx context.Context,
	appID, proposerName, nomineeName, nomineeLinkedIn, proposalReason string,
	ykMembers []struct {
		ID    string
		Email string
		Name  string
	},
) error {
	reviewURL := fmt.Sprintf("%s/applications/%s", s.baseURL, appID)
	subject := "[Membership System] Onursal Üyelik Önerisi"

	for _, yk := range ykMembers {
		data := HonoraryProposalData{
			YKMemberName:    yk.Name,
			ProposerName:    proposerName,
			NomineeName:     nomineeName,
			NomineeLinkedIn: nomineeLinkedIn,
			ProposalReason:  proposalReason,
			ReviewURL:       reviewURL,
		}

		html, text, err := Render(tmplHonoraryProposal, data)
		if err != nil {
			return err
		}

		if err := s.mailer.Send(yk.Email, subject, html, text); err != nil {
			s.logEmail(ctx, appID, "application", "email.failed", yk.Email)
			return fmt.Errorf("notifications: send honorary proposal to %s: %w", yk.Email, err)
		}

		s.logEmail(ctx, appID, "application", "email.honorary_proposal", yk.Email)
	}
	return nil
}

// ─── helpers ───────────────────────────────────────────────────────────────────

func (s *Service) logEmail(ctx context.Context, entityID, entityType, action, to string) {
	meta, _ := json.Marshal(map[string]string{"to": to})
	entry := emailLogRow{
		ID:         uuid.New().String(),
		Action:     action,
		ActorRole:  "system",
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   datatypes.JSON(meta),
		CreatedAt:  time.Now(),
	}
	_ = s.db.WithContext(ctx).Create(&entry)
}
