package gateway

// Test-only re-exports of unexported identifiers used by the external
// gateway_test package. Confined to *_test.go so it does not affect
// the public API.

type (
	StudyRecordIter        = studyRecordIter
	StudyRecordBulkDeleter = studyRecordBulkDeleter
)

func DeleteStudyRecordDocs(iter StudyRecordIter, bw StudyRecordBulkDeleter) error {
	return deleteStudyRecordDocs(iter, bw)
}
