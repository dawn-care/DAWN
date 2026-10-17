import { useStaffCommands } from "./useStaffCommands";

/** Receives staff play and stop commands while mounted. Draws nothing. */
export default function StaffCommands() {
  useStaffCommands();
  return null;
}
