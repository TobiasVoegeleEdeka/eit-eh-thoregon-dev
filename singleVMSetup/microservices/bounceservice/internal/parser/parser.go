package parser

import (
	"bounceservice/types"
	"strings"
)

func ParseBounceLine(line string) types.Bounce {
	bounce := types.Bounce{
		Raw: line,
	}

	if parts := strings.SplitN(line, " ", 3); len(parts) >= 2 {
		bounce.Timestamp = parts[0] + " " + parts[1]
	}

	bounce.QueueID = extractField(line, "queueid=", " ")
	bounce.From = extractField(line, "from=<", ">")
	bounce.To = extractField(line, "to=<", ">")
	bounce.Status = extractField(line, "status=", " ")
	bounce.Reason = extractBounceReason(line)

	if bounce.Reason == "" {
		bounce.Reason = "Unknown reason"
	}
	if bounce.From == "" {
		bounce.From = extractField(line, "from=", " ")
	}
	if bounce.To == "" {
		bounce.To = extractField(line, "to=", " ")
	}

	return bounce
}

func extractField(line, startDelim, endDelim string) string {
	startIdx := strings.Index(line, startDelim)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(startDelim)
	endIdx := strings.Index(line[startIdx:], endDelim)
	if endIdx == -1 {
		return line[startIdx:]
	}

	return line[startIdx : startIdx+endIdx]
}

func extractBounceReason(line string) string {
	if reason := extractField(line, "reason=", " "); reason != "" {
		return reason
	}

	switch {
	case strings.Contains(line, "User unknown"):
		return "Recipient address does not exist"
	case strings.Contains(line, "Mailbox full"):
		return "Recipient mailbox is full"
	case strings.Contains(line, "Relay access denied"):
		return "Relay access denied"
	case strings.Contains(line, "Blocked"):
		return "Blocked by recipient server"
	case strings.Contains(line, "host not found"):
		return "Recipient domain not found"
	case strings.Contains(line, "Message too large"):
		return "Message size exceeds limit"
	case strings.Contains(line, "spam"):
		return "Marked as spam"
	case strings.Contains(line, "rejected"):
		return "Rejected by recipient server"
	}

	if parts := strings.Split(line, "status=bounced"); len(parts) > 1 {
		reason := strings.TrimSpace(parts[1])
		if len(reason) > 100 {
			reason = reason[:100] + "..."
		}
		return reason
	}

	return ""
}
