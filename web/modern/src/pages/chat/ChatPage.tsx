export function ChatPage() {
  const chatLink = localStorage.getItem('chat_link')

  if (!chatLink) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <h2 className="text-2xl font-semibold text-muted-foreground mb-2">
            Chat Not Available
          </h2>
          <p className="text-muted-foreground">
            Chat service has not been configured by the administrator.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full">
      <iframe
        src={chatLink}
        className="w-full h-full border-0"
        title="Chat Interface"
      />
    </div>
  )
}

export default ChatPage
