import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AuthProvider } from "./index";

describe("AuthProvider", () => {
  it("renders children", () => {
    render(
      <AuthProvider>
        <span>child content</span>
      </AuthProvider>,
    );

    expect(screen.getByText("child content")).toBeDefined();
  });
});

