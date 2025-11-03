import * as React from "react"
import { cn } from "@/lib/utils"
import { Card, CardContent } from "@/components/ui/card"
import { ChevronDown, ChevronRight, CheckCircle2, XCircle, Loader2 } from "lucide-react"

interface ToolProps extends React.HTMLAttributes<HTMLDivElement> {
  status?: "idle" | "loading" | "success" | "error"
  defaultOpen?: boolean
}

const Tool = React.forwardRef<HTMLDivElement, ToolProps>(
  ({ className, status = "idle", defaultOpen = false, children, ...props }, ref) => {
    const [isOpen, setIsOpen] = React.useState(defaultOpen)

    return (
      <Card
        ref={ref}
        className={cn(
          "border-l-4 transition-colors",
          status === "success" && "border-l-green-500",
          status === "error" && "border-l-red-500",
          status === "loading" && "border-l-blue-500",
          status === "idle" && "border-l-muted-foreground",
          className
        )}
        {...props}
      >
        <CardContent className="p-0">
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="w-full text-left p-4 hover:bg-muted/50 transition-colors"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 flex-1">
                {React.Children.toArray(children).find(
                  (child) => React.isValidElement(child) && child.type === ToolHeader
                )}
              </div>
              {isOpen ? (
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              ) : (
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              )}
            </div>
          </button>
          
          {isOpen && (
            <div className="px-4 pb-4 space-y-3">
              {React.Children.toArray(children).filter(
                (child) => React.isValidElement(child) && child.type !== ToolHeader
              )}
            </div>
          )}
        </CardContent>
      </Card>
    )
  }
)
Tool.displayName = "Tool"

const ToolHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & { status?: "idle" | "loading" | "success" | "error" }
>(({ className, status, children, ...props }, ref) => {
  const StatusIcon = {
    idle: null,
    loading: Loader2,
    success: CheckCircle2,
    error: XCircle,
  }[status || "idle"]

  return (
    <div
      ref={ref}
      className={cn("flex items-center gap-2", className)}
      {...props}
    >
      {StatusIcon && (
        <StatusIcon
          className={cn(
            "h-4 w-4",
            status === "loading" && "animate-spin text-blue-500",
            status === "success" && "text-green-500",
            status === "error" && "text-red-500"
          )}
        />
      )}
      <span className="font-semibold text-sm">{children}</span>
    </div>
  )
})
ToolHeader.displayName = "ToolHeader"

const ToolContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("space-y-2", className)}
    {...props}
  >
    {children}
  </div>
))
ToolContent.displayName = "ToolContent"

const ToolInput = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("text-sm", className)}
    {...props}
  >
    <div className="font-medium text-muted-foreground mb-1">Input</div>
    <div className="bg-muted/50 rounded-md p-3 font-mono text-xs overflow-x-auto">
      {children}
    </div>
  </div>
))
ToolInput.displayName = "ToolInput"

const ToolOutput = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("text-sm", className)}
    {...props}
  >
    <div className="font-medium text-muted-foreground mb-1">Output</div>
    <div className="bg-muted/50 rounded-md p-3 text-xs overflow-x-auto">
      {children}
    </div>
  </div>
))
ToolOutput.displayName = "ToolOutput"

export { Tool, ToolHeader, ToolContent, ToolInput, ToolOutput }

