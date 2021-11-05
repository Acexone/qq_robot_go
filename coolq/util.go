package coolq

import (
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
)

// 单条消息发送的大小限制（预估）
const MaxMessageSize = 5000

// SplitLongMessage 将过长的消息分割为若干个适合发送的消息
func SplitLongMessage(sendingMessage *message.SendingMessage) []*message.SendingMessage {
	// 合并连续文本消息
	sendingMessage = mergeContinuousTextMessages(sendingMessage)

	// 分割过长元素
	sendingMessage = splitElements(sendingMessage)

	// 将元素分为多组，确保各组不超过单条消息的上限
	splitMessages := splitMessages(sendingMessage)

	return splitMessages
}

// mergeContinuousTextMessages 预先将所有连续的文本消息合并为到一起，方便后续统一切割
func mergeContinuousTextMessages(sendingMessage *message.SendingMessage) *message.SendingMessage {
	// 检查下是否有连续的文本消息，若没有，则可以直接返回
	lastIsText := false
	hasContinuousText := false
	for _, msg := range sendingMessage.Elements {
		if msg.Type() == message.Text {
			if lastIsText {
				// 有连续的文本消息，需要进行处理
				hasContinuousText = true
				break
			}

			// 遇到文本元素先存放起来，方便将连续的文本元素合并
			lastIsText = true
			continue
		} else {
			lastIsText = false
		}
	}
	if !hasContinuousText {
		return sendingMessage
	}

	// 存在连续的文本消息，需要进行合并处理
	textBuffer := strings.Builder{}
	lastIsText = false
	totalMessageCount := 0
	for _, msg := range sendingMessage.Elements {
		if msgVal, ok := msg.(*message.TextElement); ok {
			// 遇到文本元素先存放起来，方便将连续的文本元素合并
			textBuffer.WriteString(msgVal.Content)
			lastIsText = true
			continue
		}

		// 如果之前的是文本元素（可能是多个合并起来的），则在这里将其实际放入消息中
		if lastIsText {
			sendingMessage.Elements[totalMessageCount] = message.NewText(textBuffer.String())
			totalMessageCount += 1
			textBuffer.Reset()
		}
		lastIsText = false

		// 非文本元素则直接处理
		sendingMessage.Elements[totalMessageCount] = msg
		totalMessageCount += 1
	}
	// 处理最后几个元素是文本的情况
	if textBuffer.Len() != 0 {
		sendingMessage.Elements[totalMessageCount] = message.NewText(textBuffer.String())
		totalMessageCount += 1
		textBuffer.Reset()
	}
	sendingMessage.Elements = sendingMessage.Elements[:totalMessageCount]

	return sendingMessage
}

// splitElements 将原有消息的各个元素先尝试处理，如过长的文本消息按需分割为多个元素
func splitElements(sendingMessage *message.SendingMessage) *message.SendingMessage {
	// 检查下是否存在需要文本消息，若不存在，则直接返回
	needSplit := false
	for _, msg := range sendingMessage.Elements {
		if msgVal, ok := msg.(*message.TextElement); ok {
			if textNeedSplit(msgVal.Content) {
				needSplit = true
				break
			}
		}
	}
	if !needSplit {
		return sendingMessage
	}

	// 开始尝试切割
	messageParts := message.NewSendingMessage()

	for _, msg := range sendingMessage.Elements {
		switch msgVal := msg.(type) {
		case *message.TextElement:
			messageParts.Elements = append(messageParts.Elements, splitPlainMessage(msgVal.Content)...)
		default:
			messageParts.Append(msg)
		}
	}

	return messageParts
}

// splitMessages 根据大小分为多个消息进行发送
func splitMessages(sendingMessage *message.SendingMessage) []*message.SendingMessage {
	var splitMessages []*message.SendingMessage

	messagePart := message.NewSendingMessage()
	msgSize := 0
	for _, part := range sendingMessage.Elements {
		estimateSize := message.EstimateLength([]message.IMessageElement{part})
		// 若当前分消息加上新的元素后大小会超限，且已经有元素（确保不会无限循环），则开始切分为新的一个元素
		if msgSize+estimateSize > MaxMessageSize && len(messagePart.Elements) > 0 {
			splitMessages = append(splitMessages, messagePart)

			messagePart = message.NewSendingMessage()
			msgSize = 0
		}

		// 加上新的元素
		messagePart.Append(part)
		msgSize += estimateSize
	}
	// 将最后一个分片加上
	if len(messagePart.Elements) != 0 {
		splitMessages = append(splitMessages, messagePart)
	}

	return splitMessages
}

func splitPlainMessage(content string) []message.IMessageElement {
	if !textNeedSplit(content) {
		return []message.IMessageElement{message.NewText(content)}
	}

	splittedMessage := make([]message.IMessageElement, 0, (len(content)+MaxMessageSize-1)/MaxMessageSize)

	last := 0
	for runeIndex, runeValue := range content {
		// 如果加上新的这个字符后，会超出大小，则从这个字符前分一次片
		if runeIndex+len(string(runeValue))-last > MaxMessageSize {
			splittedMessage = append(splittedMessage, message.NewText(content[last:runeIndex]))
			last = runeIndex
		}
	}
	if last != len(content) {
		splittedMessage = append(splittedMessage, message.NewText(content[last:len(content)]))
	}

	return splittedMessage
}

func textNeedSplit(content string) bool {
	return len(content) > MaxMessageSize
}
