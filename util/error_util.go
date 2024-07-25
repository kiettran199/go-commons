package util

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/kiettran199/go-commons/api"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	InternalErrorReason = "INTERNAL_ERROR"
	PlatformErrorDomain = "api.aiaengine.com"
)

func ToGRPCStatus(i interface{}) *status.Status {
	switch v := i.(type) {
	case error:
		if se, ok := v.(interface {
			GRPCStatus() *status.Status
		}); ok {
			return se.GRPCStatus()
		}
		errorNumber := uuid.New().String()
		statusErr, err := status.New(codes.Internal, fmt.Sprintf("An unexpected error occurred. Please contact us for support and refer this error number: %s.", errorNumber)).WithDetails(
			&errdetails.ErrorInfo{
				Reason: InternalErrorReason,
				Domain: PlatformErrorDomain,
				Metadata: map[string]string{
					"error_number": errorNumber,
				},
			},
			&api.SuggestionInfo{
				Suggestion: fmt.Sprintf("Please contact us for support and refer this error number: %s", errorNumber),
			},
			&errdetails.DebugInfo{
				Detail:       v.Error(),
				StackEntries: getStackTrace(5, v),
			},
		)
		if err != nil {
			return ToGRPCStatus(err)
		}
		return statusErr
	default:
		return ToGRPCStatus(fmt.Errorf("%v", i))
	}
}

func getStackTrace(maxLines int, err error) []string {
	stack := debug.Stack()
	stackTrace := make([]string, 0)
	lines := bytesToLines(stack)
	// interate the stack trace from the bottom up to locate the error origin
	startIndex := len(lines) - maxLines + 1
	if startIndex < 0 {
		startIndex = 0
	}
	for i := startIndex; i < len(lines); i++ {
		stackTrace = append(stackTrace, string(lines[i]))
	}
	stackTrace = append(stackTrace, err.Error())
	return stackTrace
}

func bytesToLines(stack []byte) []string {
	var lines []string
	var line []byte
	for _, b := range stack {
		if b == '\n' {
			lines = append(lines, string(line))
			line = nil
		} else {
			line = append(line, b)
		}
	}
	if len(line) > 0 {
		lines = append(lines, string(line))
	}
	return lines
}

func fromGRPCStatusToMap(status *status.Status) map[string]interface{} {
	if status == nil {
		return nil
	}
	bytes, err := protojson.Marshal(status.Proto())
	if err != nil {
		log.Printf("Unable to parse status = %v, error = %v", status, err)
		return nil
	}
	r := map[string]interface{}{}
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		log.Printf("Unable to parse json = %v, error = %v", bytes, err)
		return nil
	}
	return r
}

func ToArgoWritableReportString(i interface{}) string {
	status := ToGRPCStatus(i)
	return FromGRPCStatusToArgoWritableReportString(status)
}

func FromGRPCStatusToArgoWritableReportString(status *status.Status) string {
	statusMap := fromGRPCStatusToMap(status)
	if statusMap != nil {
		reportStr := map[string]interface{}{
			"error": statusMap,
		}
		bytes, err := json.Marshal(reportStr)
		if err != nil {
			log.Printf("Unable to parse data: %v", err)
			return ""
		}
		return string(bytes)
	}
	return ""
}

func BuildStatusError(c codes.Code, reason string, message string) error {
	statusErr, _ := status.New(c, message).WithDetails(&errdetails.ErrorInfo{
		Reason:   reason,
		Domain:   PlatformErrorDomain,
		Metadata: nil,
	})
	return status.ErrorProto(statusErr.Proto())
}
