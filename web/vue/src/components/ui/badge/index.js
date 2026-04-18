import { cva } from "class-variance-authority";

export { default as Badge } from "./Badge.vue";

export const badgeVariants = cva(
  'tw-inline-flex tw-gap-1 tw-items-center tw-rounded-full tw-border tw-px-2.5 tw-py-0.5 tw-text-xs tw-font-semibold tw-transition-colors focus:tw-outline-none focus:tw-ring-2 focus:tw-ring-ring focus:tw-ring-offset-2',
  {
    variants: {
      variant: {
        default:
          'tw-border-transparent tw-bg-primary tw-text-primary-foreground hover:tw-bg-primary/80',
        secondary:
          'tw-border-transparent tw-bg-secondary tw-text-secondary-foreground hover:tw-bg-secondary/80',
        destructive:
          'tw-border-transparent tw-bg-destructive tw-text-destructive-foreground hover:tw-bg-destructive/80',
        outline: 'tw-text-foreground',
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);
