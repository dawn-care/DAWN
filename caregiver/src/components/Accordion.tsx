import { useState } from "react";

export default function Accordion({
  title,
  children,
  defaultOpen = false,
}: {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="border rounded-lg">
      <button
        onClick={() => setOpen(!open)}
        className="w-full text-left px-6 py-4 flex justify-between items-center bg-gray-100 hover:bg-gray-200 rounded-lg"
      >
        <span className="text-xl font-semibold">{title}</span>
        <span className="text-2xl">{open ? "−" : "+"}</span>
      </button>

      {open && (
        <div className="px-6 py-4 space-y-4 bg-white">
          {children}
        </div>
      )}
    </div>
  );
}