import { useAuth } from "../context/AuthContext";
import { Navigate } from "react-router-dom";
import type { JSX } from "react";

export default function RequireCarePartner({ children }: { children: JSX.Element }) {
  const { activeCarePartnerId } = useAuth();

  if (!activeCarePartnerId) {
    return <Navigate to="/setup" replace />;
  }

  return children;
}