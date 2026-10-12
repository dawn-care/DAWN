/**
 * RadioGroup.tsx
 *
 * User: Caregiver
 *
 * A group of radio options, built on Radix so it has real radio
 * semantics (role="radiogroup", arrow-key navigation, one tab stop).
 * Put RadioOption children inside it. Always pass an aria-label.
 */

import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import type { ComponentPropsWithoutRef } from "react";

type RadioGroupProps = ComponentPropsWithoutRef<typeof RadixRadioGroup.Root>;

export default function RadioGroup({ className = "", ...rootProps }: RadioGroupProps) {
  return <RadixRadioGroup.Root className={`flex flex-col gap-1.5 ${className}`} {...rootProps} />;
}
