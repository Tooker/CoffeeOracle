import nextJest from "next/jest.js";

const createJestConfig = nextJest({ dir: "./" });

// config defines the test runtime (jsdom), setup hooks, and path/style mapping.
const config = {
  coverageProvider: "v8",
  setupFilesAfterEnv: ["<rootDir>/jest.setup.ts"],
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/$1",
    "^.+\\.(css|less|scss|sass)$": "identity-obj-proxy",
  },
  testEnvironment: "jest-environment-jsdom",
};

export default createJestConfig(config);
