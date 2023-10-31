// Code generated by woco, DO NOT EDIT.

package file

// DeleteFileRequest is the request object for (DELETE /files/{fileId})
type DeleteFileRequest struct {
	FileId string `binding:"required" uri:"fileId"`
}

// GetFileRequest is the request object for (GET /files/{fileId})
type GetFileRequest struct {
	FileId string `binding:"required" uri:"fileId"`
}

// GetFileRawRequest is the request object for (GET /files/{fileId}/raw)
type GetFileRawRequest struct {
	FileId string `binding:"required" uri:"fileId"`
}

// ReportRefCountRequest is the request object for (POST /files/report-ref-count)
type ReportRefCountRequest struct {
	Inputs []*FileRefInput `binding:"required" json:"inputs"`
}

// UploadFileRequest is the request object for (POST /files)
type UploadFileRequest struct {
	// Bucket the name of bucket，value is local
	Bucket string `form:"bucket"`
	// Key the path of file in the bucket
	Key string `form:"key"`
}

// UploadFileInfoRequest is the request object for (POST /files/upload-info)
type UploadFileInfoRequest struct {
	File       FileInput       `json:"file"`
	FileSource FileSourceInput `json:"fileSource"`
}
