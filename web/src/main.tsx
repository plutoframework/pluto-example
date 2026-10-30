import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { SessionProvider } from "@/features/session/public";
import { createQueryClient } from "./app/query-client";
import { router } from "./app/router";
import "./app/styles.css";

const element = document.getElementById("root");

if (!element) {
  throw new Error("Application root is missing");
}

const client = createQueryClient();
createRoot(element).render(
  <StrictMode>
    <QueryClientProvider client={client}>
      <SessionProvider>
        <RouterProvider router={router} />
      </SessionProvider>
    </QueryClientProvider>
  </StrictMode>,
);
