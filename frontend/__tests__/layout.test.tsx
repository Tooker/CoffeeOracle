import React from "react";
import { render, screen } from "@testing-library/react";
import Home from "@/app/page";

// Mock OracleExperience so this test focuses only on page composition, not stream logic.
jest.mock("@/components/OracleExperience", () => ({
  OracleExperience: () => <div data-testid="oracle-experience" />,
}));

// Mock UI library components to keep rendering deterministic in unit tests.
jest.mock("flowbite-react", () => ({
  Button: ({ children }: { children: React.ReactNode }) => (
    <button>{children}</button>
  ),
  FileInput: ({
    onChange,
  }: {
    onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void;
  }) => <input type="file" onChange={onChange} />,
  Label: ({ children }: { children: React.ReactNode }) => <label>{children}</label>,
  TextInput: ({
    onChange,
    value,
  }: {
    onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void;
    value?: string;
  }) => <input value={value ?? ""} onChange={onChange} />,
  RangeSlider: ({
    onChange,
    value,
  }: {
    onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void;
    value?: number;
  }) => <input type="range" value={value ?? 0} onChange={onChange} />,
}));

describe("Home", () => {
  // Verifies that the landing page CTA headline is visible.
  it("renders hero call to action", () => {
    render(<Home />);
    expect(screen.getByText(/coffee oracle/i)).toBeInTheDocument();
  });
});
