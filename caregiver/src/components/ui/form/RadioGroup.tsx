/**
 * RadioGroup.tsx
 *
 * User: Caregiver
 *
 * A group of radio options, built on Radix so it has real radio
 * semantics (role="radiogroup", arrow-key navigation, one tab stop).
 * Put RadioOption children inside it. A name is required: pass either
 * aria-label or aria-labelledby.
 */

import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import type { ComponentPropsWithoutRef } from "react";

type RadioGroupProps = ComponentPropsWithoutRef<typeof RadixRadioGroup.Root> &
  ({ "aria-label": string } | { "aria-labelledby": string });

export default function RadioGroup({ className = "", ...rootProps }: RadioGroupProps) {
  return <RadixRadioGroup.Root className={`flex flex-col gap-1.5 ${className}`} {...rootProps} />;
}
