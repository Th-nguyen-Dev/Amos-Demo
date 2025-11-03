import * as React from "react"
import { Tool, ToolHeader, ToolContent, ToolInput, ToolOutput } from "./tool"

interface ToolCall {
  name: string
  input: Record<string, string | number | boolean>
  output?: string
  status: "loading" | "success" | "error"
}

interface MessageContentProps {
  content: string
}

// Parse tool calls from the formatted markdown output
function parseToolCalls(content: string): { text: string; tools: ToolCall[] } {
  const tools: ToolCall[] = []

  // Match tool call blocks
  const toolPattern = /─{50}\n🔧 \*\*Tool Call: (.+?)\*\*\n\n📋 \*\*Arguments:\*\*\n([\s\S]*?)\n⏳ Executing\.\.\.\n([\s\S]*?)(?:─{50}|$)/g
  
  let match
  let lastIndex = 0
  const textParts: string[] = []

  while ((match = toolPattern.exec(content)) !== null) {
    // Add text before this tool call
    if (match.index > lastIndex) {
      textParts.push(content.substring(lastIndex, match.index))
    }

    const toolName = match[1]
    const argsSection = match[2]
    const resultsSection = match[3]

    // Parse arguments
    const input: Record<string, string | number | boolean> = {}
    const argMatches = argsSection.matchAll(/• \*\*(.+?):\*\* (.+?)(?=\n|$)/g)
    for (const argMatch of argMatches) {
      input[argMatch[1]] = argMatch[2]
    }

    // Determine status from results section
    let status: "loading" | "success" | "error" = "loading"
    let output = ""

    if (resultsSection.includes("✅ **Status:** Success")) {
      status = "success"
    } else if (resultsSection.includes("❌ **Status:** No Results")) {
      status = "error"
    }

    // Extract output preview
    const outputMatch = resultsSection.match(/📤 \*\*Result Preview:\*\*\n```\n([\s\S]*?)\n```/)
    if (outputMatch) {
      output = outputMatch[1]
    }

    tools.push({ name: toolName, input, output, status })
    lastIndex = toolPattern.lastIndex
  }

  // Add remaining text
  if (lastIndex < content.length) {
    textParts.push(content.substring(lastIndex))
  }

  // Clean up the text parts (remove tool markers that weren't matched)
  const cleanedText = textParts
    .join("")
    .replace(/─{50}\n/g, "")
    .replace(/🔧 \*\*Tool Call:.*?\*\*\n/g, "")
    .replace(/📋 \*\*Arguments:\*\*\n/g, "")
    .replace(/⏳ Executing\.\.\.\n/g, "")
    .replace(/✅ \*\*Status:.*?\n/g, "")
    .replace(/❌ \*\*Status:.*?\n/g, "")
    .replace(/📤 \*\*Result Preview:\*\*\n```[\s\S]*?```\n/g, "")
    .replace(/💡 Trying alternative approach\.\.\.\n/g, "")
    .trim()

  return { text: cleanedText, tools }
}

export function MessageContent({ content }: MessageContentProps) {
  const { text, tools } = React.useMemo(() => parseToolCalls(content), [content])

  return (
    <div className="space-y-4">
      {/* Display tool calls */}
      {tools.length > 0 && (
        <div className="space-y-2">
          {tools.map((tool, index) => (
            <Tool key={index} status={tool.status} defaultOpen={index === tools.length - 1}>
              <ToolHeader status={tool.status}>
                🔧 {tool.name}
              </ToolHeader>
              <ToolContent>
                {Object.keys(tool.input).length > 0 && (
                  <ToolInput>
                    <pre className="whitespace-pre-wrap">
                      {JSON.stringify(tool.input, null, 2)}
                    </pre>
                  </ToolInput>
                )}
                {tool.output && (
                  <ToolOutput>
                    <div className="whitespace-pre-wrap">{tool.output}</div>
                  </ToolOutput>
                )}
              </ToolContent>
            </Tool>
          ))}
        </div>
      )}

      {/* Display the cleaned response text */}
      {text && (
        <div className="prose prose-sm max-w-none dark:prose-invert">
          <div className="whitespace-pre-wrap text-sm leading-relaxed">
            {text}
          </div>
        </div>
      )}
    </div>
  )
}

