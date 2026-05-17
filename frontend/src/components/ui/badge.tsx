/* eslint-disable react-refresh/only-export-components */
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-mono font-[400] tracking-wider uppercase whitespace-nowrap transition-colors border",
  {
    variants: {
      variant: {
        default: "bg-canvas-soft text-body-mid border-transparent",
        secondary: "bg-canvas-soft text-body border-hairline",
        destructive: "bg-destructive/10 text-destructive border-destructive/20",
        success: "bg-success/10 text-success border-success/20",
        outline: "bg-transparent text-body border-hairline",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

function Badge({
  className,
  variant = "default",
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badgeVariants>) {
  return (
    <span className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
