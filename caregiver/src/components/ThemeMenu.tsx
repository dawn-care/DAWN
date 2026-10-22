/**
 * ThemeMenu.tsx
 *
 * User: Caregiver
 *
 * An icon button in the nav bar that opens a small menu to choose Light, Dark
 * or System. The choice is saved by theme/theme.ts and applied at once.
 */

import { useState } from "react";
import * as Popover from "@radix-ui/react-popover";
import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import { Monitor, Moon, Sun } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import IconButton from "./ui/IconButton";
import { FOCUS_RING } from "./ui/styles";
import { applyTheme, isThemeChoice, loadTheme } from "../theme/theme";
import type { ThemeChoice } from "../theme/theme";

const OPTIONS: { value: ThemeChoice; label: string; icon: LucideIcon }[] = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor },
];

export default function ThemeMenu() {
  const [choice, setChoice] = useState<ThemeChoice>(loadTheme);
  const CurrentIcon = OPTIONS.find((option) => option.value === choice)?.icon ?? Monitor;

  function handleChange(value: string) {
    if (!isThemeChoice(value)) return;
    setChoice(value);
    applyTheme(value);
  }

  return (
    <Popover.Root>
      <Popover.Trigger asChild>
        <IconButton
          label="Color theme"
          icon={<CurrentIcon size={20} aria-hidden="true" />}
          className="text-on-nav hover:bg-on-nav/10 data-[state=open]:bg-on-nav/15"
        />
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          side="top"
          align="start"
          sideOffset={8}
          className="z-50 w-[180px] p-1.5 bg-surface-raised border-[1.5px] border-border rounded-xl shadow-popover text-ink"
        >
          <RadixRadioGroup.Root
            value={choice}
            onValueChange={handleChange}
            aria-label="Color theme"
            className="flex flex-col gap-1"
          >
            {OPTIONS.map(({ value, label, icon: Icon }) => (
              <RadixRadioGroup.Item
                key={value}
                value={value}
                className={`flex items-center gap-2.5 h-11 px-2.5 rounded-lg border-0 bg-transparent text-left text-base cursor-pointer hover:bg-border/60 data-[state=checked]:bg-border data-[state=checked]:font-bold ${FOCUS_RING}`}
              >
                <Icon size={20} aria-hidden="true" />
                {label}
              </RadixRadioGroup.Item>
            ))}
          </RadixRadioGroup.Root>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
