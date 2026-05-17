/* eslint-disable react-refresh/only-export-components */
import { Button as ButtonPrimitive } from "@base-ui/react/button";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "group/button inline-flex shrink-0 items-center justify-center border text-sm font-[400] whitespace-nowrap transition-all duration-150 outline-none select-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/30 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground border-primary hover:opacity-90",
        outline:
          "bg-transparent text-ink border-hairline hover:bg-canvas-soft hover:text-ink",
        destructive:
          "bg-transparent text-destructive border-destructive/30 hover:bg-destructive/10",
        ghost:
          "bg-transparent text-body-mid border-transparent hover:bg-canvas-soft hover:text-ink",
      },
      size: {
        default: "h-8 gap-1.5 px-4 rounded-full",
        sm: "h-7 gap-1 px-3 text-xs rounded-full",
        lg: "h-9 gap-1.5 px-5 rounded-full",
        icon: "h-8 w-8 rounded-full",
        "icon-sm": "h-7 w-7 rounded-full",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

function Button({
  className,
  variant = "default",
  size = "default",
  ...props
}: ButtonPrimitive.Props & VariantProps<typeof buttonVariants>) {
  return (
    <ButtonPrimitive
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  );
}

export { Button, buttonVariants };
