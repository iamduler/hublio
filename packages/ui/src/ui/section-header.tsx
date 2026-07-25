import * as React from "react"
import { cn } from "../lib/utils"

function Eyebrow({ className, ...props }: React.HTMLAttributes<HTMLSpanElement>) {
  return <span className={cn("eyebrow", className)} {...props} />
}

function SectionHeader({
  eyebrow,
  title,
  lede,
  align = "left",
  className,
  children}: {
  eyebrow?: React.ReactNode
  title: React.ReactNode
  lede?: React.ReactNode
  align?: "left" | "center"
  className?: string
  children?: React.ReactNode
}) {
  return (
    <div className={cn(align === "center" && "text-center", className)}>
      {eyebrow ? <Eyebrow>{eyebrow}</Eyebrow> : null}
      <h2 className="sec-title">{title}</h2>
      {lede ? <p className={cn("sec-lede", align === "center" && "mx-auto")}>{lede}</p> : null}
      {children}
    </div>
  )
}

export { Eyebrow, SectionHeader }
