package article

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupE2ETestRouter() (*gin.Engine, *Service) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	// Test identities stand in for the production Firebase middleware.
	engine.Use(func(c *gin.Context) {
		c.Set("user_id", c.GetHeader("X-User-ID"))
		c.Set("username", c.GetHeader("X-User-Name"))
		c.Set("user_role", c.GetHeader("X-User-Role"))
		c.Next()
	})

	repo := newArticleRepoStub()
	svc := NewService(repo)
	handler := NewHandler(svc)

	api := engine.Group("/api")
	RegisterRoutes(api, handler)

	apiV1 := engine.Group("/api/v1")
	RegisterRoutes(apiV1, handler)

	return engine, svc
}

// TC-ALC-001 [Happy Path]: Writer สร้าง Draft บทความใหม่พร้อมฟิลด์ครบถ้วน (Full Fields)
func TestE2E_TC_ALC_001_WriterCreateDraftFullFields(t *testing.T) {
	router, _ := setupE2ETestRouter()

	payload := map[string]interface{}{
		"title":              "สุนทรียศาสตร์แห่งตัวอักษร: การเดินทางของฟอนต์ไทยร่วมสมัย",
		"slug":               "contemporary-thai-typography-journey",
		"subtitle":           "วิวัฒนาการจากบล็อกไม้พิมพ์สู่อักขระดิจิทัล",
		"excerpt":            "สำรวจรากเหง้าและพลวัตของตัวอักษรไทยในยุคเรเนซองส์แห่งการออกแบบสื่อสาร",
		"content":            "<h2>กำเนิดแห่งเส้นสาย</h2><p>ตัวอักษรไทยไม่ได้เป็นเพียงเครื่องมือบันทึกภาษา แต่เป็นสถาปัตยกรรมทางสายตา...</p>",
		"cover_image_url":    "https://cdn.artsgoz.com/uploads/2026/typography-cover.webp",
		"writer_credit":      "วริศรา อักษรสาส์น",
		"proofreader_credit": "กมลชนก ตรวจทาน",
		"illustrator_credit": "ธนพล วาดเส้น",
		"meta_title":         "สุนทรียศาสตร์แห่งตัวอักษรไทยร่วมสมัย | Artsgoz x AksornSara",
		"meta_description":   "เจาะลึกพัฒนาการและการเปลี่ยนผ่านของงานออกแบบตัวอักษรไทยในโลกยุคดิจิทัล",
		"categories": []map[string]interface{}{
			{"category_id": 101, "name": "ศิลปะและการออกแบบ", "is_primary": true},
			{"category_id": 105, "name": "วรรณกรรมวิจารณ์", "is_primary": false},
		},
		"tags":       []string{"typography", "graphic-design", "thai-art"},
		"status":     "draft",
		"visibility": "public",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer mock.jwt.writer")
	req.Header.Set("X-User-ID", "w-1111-1111")
	req.Header.Set("X-User-Name", "วริศรา อักษรสาส์น")
	req.Header.Set("X-User-Role", "Writer")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("[TC-ALC-001] expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var res Article
	var envelope struct {
		Success bool    `json:"success"`
		Data    Article `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err == nil && envelope.Data.ID != "" {
		res = envelope.Data
	} else if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("[TC-ALC-001] JSON unmarshal error: %v", err)
	}

	if res.ID == "" {
		t.Errorf("[TC-ALC-001] expected generated ID, got empty")
	}
	if res.Status != StatusDraft {
		t.Errorf("[TC-ALC-001] expected status 'draft', got %q", res.Status)
	}
	if res.AuthorID != "w-1111-1111" {
		t.Errorf("[TC-ALC-001] expected author_id 'w-1111-1111', got %q", res.AuthorID)
	}
	if res.PublishedAt != nil {
		t.Errorf("[TC-ALC-001] expected published_at to be nil for draft, got %v", res.PublishedAt)
	}
	if res.PublishedBy != nil {
		t.Errorf("[TC-ALC-001] expected published_by to be nil for draft, got %v", res.PublishedBy)
	}
	if res.WriterCredit != "วริศรา อักษรสาส์น" || res.ProofreaderCredit != "กมลชนก ตรวจทาน" || res.IllustratorCredit != "ธนพล วาดเส้น" {
		t.Errorf("[TC-ALC-001] credit fields mismatch: writer=%q, proofreader=%q, illustrator=%q",
			res.WriterCredit, res.ProofreaderCredit, res.IllustratorCredit)
	}
	if res.MetaTitle != "สุนทรียศาสตร์แห่งตัวอักษรไทยร่วมสมัย | Artsgoz x AksornSara" {
		t.Errorf("[TC-ALC-001] expected meta_title preserved, got %q", res.MetaTitle)
	}
}

// TC-ALC-002 [Negative Test]: ส่งค่า Slug ซ้ำกับบทความที่มีอยู่เดิม
func TestE2E_TC_ALC_002_DuplicateSlugConflict(t *testing.T) {
	router, _ := setupE2ETestRouter()

	// First article
	p1 := map[string]interface{}{
		"title":   "บทความแรก",
		"slug":    "duplicate-slug-test",
		"content": "เนื้อหาบทความแรก",
		"status":  "draft",
	}
	b1, _ := json.Marshal(p1)
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-User-ID", "w-1111")
	req1.Header.Set("X-User-Role", "Writer")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("[TC-ALC-002 Precondition] failed to create first article: %d", rec1.Code)
	}

	// Second article with identical slug
	p2 := map[string]interface{}{
		"title":   "บทความที่สองส่ง Slug ซ้ำ",
		"slug":    "duplicate-slug-test",
		"content": "เนื้อหาพยายามใช้ slug ซ้ำ",
		"status":  "draft",
	}
	b2, _ := json.Marshal(p2)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-User-ID", "w-1111")
	req2.Header.Set("X-User-Role", "Writer")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusConflict && rec2.Code != http.StatusBadRequest {
		t.Fatalf("[TC-ALC-002] expected 409 Conflict or 400 Bad Request, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

// TC-ALC-003 [Negative Test]: ตรวจสอบ Partial Unique Index: Category หลัก (is_primary = true) ซ้ำกัน 2 รายการ
func TestE2E_TC_ALC_003_MultiplePrimaryCategoriesConstraint(t *testing.T) {
	router, _ := setupE2ETestRouter()

	payload := map[string]interface{}{
		"title":   "บทความตั้ง Category หลักซ้ำซ้อน",
		"slug":    "multiple-primary-categories-test",
		"content": "ทดสอบ Partial Unique Index",
		"categories": []map[string]interface{}{
			{"category_id": 101, "is_primary": true},
			{"category_id": 102, "is_primary": true},
		},
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "w-1111")
	req.Header.Set("X-User-Role", "Writer")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("[TC-ALC-003] expected 400 Bad Request for duplicate primary category, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TC-RBAC-001 [Happy Path]: Writer แก้ไขเนื้อหาบทความของตนเอง
func TestE2E_TC_RBAC_001_WriterUpdateOwnArticle(t *testing.T) {
	router, _ := setupE2ETestRouter()

	// Create article as Writer A
	p := map[string]interface{}{
		"title":   "บทความของ Writer A",
		"content": "เนื้อหาเดิม",
		"status":  "draft",
	}
	b, _ := json.Marshal(p)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-ID", "w-1111-1111")
	reqCreate.Header.Set("X-User-Role", "Writer")
	recCreate := httptest.NewRecorder()
	router.ServeHTTP(recCreate, reqCreate)

	var created Article
	_ = json.Unmarshal(recCreate.Body.Bytes(), &created)

	// Writer A updates their own article
	updatePayload := map[string]interface{}{
		"title":   "บทความของ Writer A (แก้ไขสำเร็จ)",
		"content": "เนื้อหาที่อัปเดตแล้ว",
	}
	bUp, _ := json.Marshal(updatePayload)
	reqUp, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/articles/"+created.ID, bytes.NewReader(bUp))
	reqUp.Header.Set("Content-Type", "application/json")
	reqUp.Header.Set("X-User-ID", "w-1111-1111")
	reqUp.Header.Set("X-User-Role", "Writer")

	recUp := httptest.NewRecorder()
	router.ServeHTTP(recUp, reqUp)

	if recUp.Code != http.StatusOK {
		t.Fatalf("[TC-RBAC-001] expected 200 OK, got %d: %s", recUp.Code, recUp.Body.String())
	}

	var updated Article
	_ = json.Unmarshal(recUp.Body.Bytes(), &updated)
	if updated.Title != "บทความของ Writer A (แก้ไขสำเร็จ)" {
		t.Errorf("[TC-RBAC-001] title not updated: %q", updated.Title)
	}
}

// TC-RBAC-002 [Negative Test - BOLA/IDOR]: Writer B พยายามแก้ไขบทความของ Writer A
func TestE2E_TC_RBAC_002_WriterCannotUpdateOtherWriterArticle(t *testing.T) {
	router, _ := setupE2ETestRouter()

	// Create article as Writer A
	p := map[string]interface{}{
		"title":   "บทความสงวนสิทธิ์ของ Writer A",
		"content": "เนื้อหาต้นฉบับ",
		"status":  "draft",
	}
	b, _ := json.Marshal(p)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-ID", "w-1111-1111")
	reqCreate.Header.Set("X-User-Role", "Writer")
	recCreate := httptest.NewRecorder()
	router.ServeHTTP(recCreate, reqCreate)

	var created Article
	_ = json.Unmarshal(recCreate.Body.Bytes(), &created)

	// Writer B attempts to update Writer A's article
	hackPayload := map[string]interface{}{
		"title":   "ถูกแฮกโดย Writer B",
		"content": "พยายาม BOLA",
	}
	bHack, _ := json.Marshal(hackPayload)
	reqHack, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/articles/"+created.ID, bytes.NewReader(bHack))
	reqHack.Header.Set("Content-Type", "application/json")
	reqHack.Header.Set("X-User-ID", "w-2222-2222")
	reqHack.Header.Set("X-User-Role", "Writer")

	recHack := httptest.NewRecorder()
	router.ServeHTTP(recHack, reqHack)

	if recHack.Code != http.StatusForbidden {
		t.Fatalf("[TC-RBAC-002] expected 403 Forbidden for BOLA attempt, got %d: %s", recHack.Code, recHack.Body.String())
	}
}

// TC-RBAC-003 [Negative Test - BFLA]: Writer พยายาม Publish บทความโดยตรง
func TestE2E_TC_RBAC_003_WriterCannotPublishDirectly(t *testing.T) {
	router, _ := setupE2ETestRouter()

	p := map[string]interface{}{
		"title":   "บทความรอการตรวจสอบ",
		"content": "เนื้อหา",
		"status":  "draft",
	}
	b, _ := json.Marshal(p)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-ID", "w-1111-1111")
	reqCreate.Header.Set("X-User-Role", "Writer")
	recCreate := httptest.NewRecorder()
	router.ServeHTTP(recCreate, reqCreate)

	var created Article
	_ = json.Unmarshal(recCreate.Body.Bytes(), &created)

	// Writer attempts to PATCH status to published
	statusPayload := map[string]interface{}{
		"status": "published",
	}
	bStatus, _ := json.Marshal(statusPayload)
	reqStatus, _ := http.NewRequest(http.MethodPatch, "/api/v1/admin/articles/"+created.ID+"/status", bytes.NewReader(bStatus))
	reqStatus.Header.Set("Content-Type", "application/json")
	reqStatus.Header.Set("X-User-ID", "w-1111-1111")
	reqStatus.Header.Set("X-User-Role", "Writer")

	recStatus := httptest.NewRecorder()
	router.ServeHTTP(recStatus, reqStatus)

	if recStatus.Code != http.StatusForbidden {
		t.Fatalf("[TC-RBAC-003] expected 403 Forbidden for Writer direct publish, got %d: %s", recStatus.Code, recStatus.Body.String())
	}
}

// TC-RBAC-004 [Negative Test]: ผู้ใช้ทั่วไป (ไม่มี CMS Role) เรียก Endpoint จัดการบทความ
func TestE2E_TC_RBAC_004_RegularUserCannotAccessAdminEndpoint(t *testing.T) {
	router, _ := setupE2ETestRouter()

	p := map[string]interface{}{
		"title":   "ผู้ใช้ทั่วไปพยายามสร้างบทความ",
		"content": "ไม่มีสิทธิ์ CMS",
	}
	b, _ := json.Marshal(p)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "u-member-0001")
	req.Header.Set("X-User-Role", "Member")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("[TC-RBAC-004] expected 403 Forbidden for regular member accessing CMS, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TC-RBAC-005 [Happy Path]: Editor เข้าถึงและแก้ไขบทความของ Writer
func TestE2E_TC_RBAC_005_EditorCanUpdateAnyArticle(t *testing.T) {
	router, _ := setupE2ETestRouter()

	p := map[string]interface{}{
		"title":   "บทความเขียนโดย Writer",
		"content": "เนื้อหาเดิมจาก Writer",
		"status":  "draft",
	}
	b, _ := json.Marshal(p)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-ID", "w-1111-1111")
	reqCreate.Header.Set("X-User-Role", "Writer")
	recCreate := httptest.NewRecorder()
	router.ServeHTTP(recCreate, reqCreate)

	var created Article
	_ = json.Unmarshal(recCreate.Body.Bytes(), &created)

	// Editor edits the article
	editorPayload := map[string]interface{}{
		"title":              "บทความเขียนโดย Writer (ได้รับการรีวิวโดย Editor แล้ว)",
		"proofreader_credit": "บรรณาธิการใหญ่ ตรวจทาน",
	}
	bEdit, _ := json.Marshal(editorPayload)
	reqEdit, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/articles/"+created.ID, bytes.NewReader(bEdit))
	reqEdit.Header.Set("Content-Type", "application/json")
	reqEdit.Header.Set("X-User-ID", "e-9999-9999")
	reqEdit.Header.Set("X-User-Role", "Editor")

	recEdit := httptest.NewRecorder()
	router.ServeHTTP(recEdit, reqEdit)

	if recEdit.Code != http.StatusOK {
		t.Fatalf("[TC-RBAC-005] expected 200 OK for Editor update, got %d: %s", recEdit.Code, recEdit.Body.String())
	}

	var updated Article
	var envelopeEdit struct {
		Success bool    `json:"success"`
		Data    Article `json:"data"`
	}
	if err := json.Unmarshal(recEdit.Body.Bytes(), &envelopeEdit); err == nil && envelopeEdit.Data.ID != "" {
		updated = envelopeEdit.Data
	} else {
		_ = json.Unmarshal(recEdit.Body.Bytes(), &updated)
	}
	if updated.Title != "บทความเขียนโดย Writer (ได้รับการรีวิวโดย Editor แล้ว)" {
		t.Errorf("[TC-RBAC-005] title mismatch: %q", updated.Title)
	}
	if updated.ProofreaderCredit != "บรรณาธิการใหญ่ ตรวจทาน" {
		t.Errorf("[TC-RBAC-005] proofreader credit mismatch: %q", updated.ProofreaderCredit)
	}
}

// TC-PUB-001 [Happy Path]: Editor Publish บทความ และตรวจสอบ published_at, published_by
func TestE2E_TC_PUB_001_EditorPublishArticle(t *testing.T) {
	router, _ := setupE2ETestRouter()

	p := map[string]interface{}{
		"title":   "บทความเตรียมพร้อมเผยแพร่",
		"content": "เนื้อหาพร้อมใช้งาน",
		"status":  "draft",
	}
	b, _ := json.Marshal(p)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/articles", bytes.NewReader(b))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-ID", "w-1111-1111")
	reqCreate.Header.Set("X-User-Role", "Writer")
	recCreate := httptest.NewRecorder()
	router.ServeHTTP(recCreate, reqCreate)

	var created Article
	_ = json.Unmarshal(recCreate.Body.Bytes(), &created)

	// Editor publishes
	statusPayload := map[string]interface{}{
		"status": "published",
	}
	bStatus, _ := json.Marshal(statusPayload)
	reqPub, _ := http.NewRequest(http.MethodPatch, "/api/v1/admin/articles/"+created.ID+"/status", bytes.NewReader(bStatus))
	reqPub.Header.Set("Content-Type", "application/json")
	reqPub.Header.Set("X-User-ID", "e-9999-9999")
	reqPub.Header.Set("X-User-Role", "Editor")

	recPub := httptest.NewRecorder()
	router.ServeHTTP(recPub, reqPub)

	if recPub.Code != http.StatusOK {
		t.Fatalf("[TC-PUB-001] expected 200 OK for publish, got %d: %s", recPub.Code, recPub.Body.String())
	}

	var published Article
	var envelopePub struct {
		Success bool    `json:"success"`
		Data    Article `json:"data"`
	}
	if err := json.Unmarshal(recPub.Body.Bytes(), &envelopePub); err == nil && envelopePub.Data.ID != "" {
		published = envelopePub.Data
	} else {
		_ = json.Unmarshal(recPub.Body.Bytes(), &published)
	}
	if published.Status != StatusPublished {
		t.Errorf("[TC-PUB-001] expected status 'published', got %q", published.Status)
	}
	if published.PublishedAt == nil {
		t.Errorf("[TC-PUB-001] expected published_at to be automatically set, got nil")
	}
	if published.PublishedBy == nil || *published.PublishedBy != "e-9999-9999" {
		t.Errorf("[TC-PUB-001] expected published_by to be 'e-9999-9999', got %v", published.PublishedBy)
	}
}

// TC-PUB-002 [Happy Path]: สาธารณะเรียกดูบทความที่ Publish แล้วผ่าน Public API
func TestE2E_TC_PUB_002_PublicCanViewPublishedArticle(t *testing.T) {
	router, svc := setupE2ETestRouter()

	created, err := svc.Create(t.Context(), CreateArticleInput{
		Title:   "บทความเปิดเผยแพร่สาธารณะ",
		Slug:    "public-live-article",
		Content: "เนื้อหาที่ทุกคนอ่านได้",
		Status:  "published",
	})
	if err != nil {
		t.Fatalf("[TC-PUB-002 Precondition] create error: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles/"+created.Slug, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-PUB-002] expected 200 OK for published article, got %d: %s", rec.Code, rec.Body.String())
	}

	var res Article
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Title != "บทความเปิดเผยแพร่สาธารณะ" {
		t.Errorf("[TC-PUB-002] title mismatch: %q", res.Title)
	}
}

// TC-PUB-003 [Negative Test]: สาธารณะพยายามเข้าถึงบทความฉบับร่าง (Draft) ผ่าน Public API
func TestE2E_TC_PUB_003_PublicCannotViewDraftArticle(t *testing.T) {
	router, svc := setupE2ETestRouter()

	created, err := svc.Create(t.Context(), CreateArticleInput{
		Title:   "บทความร่างลับ",
		Slug:    "secret-draft-article",
		Content: "เนื้อหาลับที่ยังไม่เผยแพร่",
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("[TC-PUB-003 Precondition] create error: %v", err)
	}

	// Public calls via slug
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles/"+created.Slug, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("[TC-PUB-003] expected 404 Not Found for draft article accessed publicly, got %d: %s", rec.Code, rec.Body.String())
	}

	// Public calls via ID
	reqID, _ := http.NewRequest(http.MethodGet, "/api/v1/articles/"+created.ID, nil)
	recID := httptest.NewRecorder()
	router.ServeHTTP(recID, reqID)

	if recID.Code != http.StatusNotFound {
		t.Fatalf("[TC-PUB-003] expected 404 Not Found for draft article accessed by ID publicly, got %d: %s", recID.Code, recID.Body.String())
	}
}

// TC-PUB-004 [Negative Test]: สาธารณะพยายามเข้าถึงบทความที่ถูกเก็บถาวร (Archived) ผ่าน Public API
func TestE2E_TC_PUB_004_PublicCannotViewArchivedArticle(t *testing.T) {
	router, svc := setupE2ETestRouter()

	created, err := svc.Create(t.Context(), CreateArticleInput{
		Title:   "บทความเก่าเก็บถาวร",
		Slug:    "old-archived-article",
		Content: "เนื้อหาที่ถูก archive แล้ว",
		Status:  "archived",
	})
	if err != nil {
		t.Fatalf("[TC-PUB-004 Precondition] create error: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles/"+created.Slug, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("[TC-PUB-004] expected 404 Not Found for archived article accessed publicly, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TC-PUB-005 [Happy Path]: Public Feed แสดงเฉพาะบทความที่มีสถานะ 'published' เท่านั้น
func TestE2E_TC_PUB_005_PublicFeedContainsOnlyPublishedArticles(t *testing.T) {
	router, svc := setupE2ETestRouter()

	// 1. Published article
	_, _ = svc.Create(t.Context(), CreateArticleInput{
		Title:   "บทความที่ 1 เผยแพร่แล้ว",
		Content: "เนื้อหา 1",
		Status:  "published",
	})

	// 2. Draft article
	_, _ = svc.Create(t.Context(), CreateArticleInput{
		Title:   "บทความที่ 2 เป็นร่าง",
		Content: "เนื้อหา 2",
		Status:  "draft",
	})

	// 3. Archived article
	_, _ = svc.Create(t.Context(), CreateArticleInput{
		Title:   "บทความที่ 3 เก็บถาวร",
		Content: "เนื้อหา 3",
		Status:  "archived",
	})

	// Anonymous / Public request
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[TC-PUB-005] expected 200 OK for public feed, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Data []Article `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("[TC-PUB-005] JSON unmarshal error: %v", err)
	}

	if len(res.Data) != 1 {
		t.Fatalf("[TC-PUB-005] expected exactly 1 published article in feed, got %d", len(res.Data))
	}

	for _, a := range res.Data {
		if a.Status != StatusPublished {
			t.Errorf("[TC-PUB-005] expected only 'published' status in public feed, found %q in article %s", a.Status, a.Title)
		}
	}
}
