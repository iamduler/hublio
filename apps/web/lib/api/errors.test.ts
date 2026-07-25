import { describe, expect, it } from "vitest";
import { ApiError } from "./client";
import { getApiErrorMessage } from "./errors";

const t = (key: string) => `t:${key}`;

describe("getApiErrorMessage", () => {
  const cases: Array<{
    name: string;
    error: unknown;
    expected: string;
  }> = [
    {
      name: "maps 401 to unauthorized",
      error: new ApiError(401, "UNAUTHORIZED", "nope"),
      expected: "t:unauthorized",
    },
    {
      name: "maps 5xx to generic",
      error: new ApiError(500, "INTERNAL", "boom"),
      expected: "t:generic",
    },
    {
      name: "returns domain message for 4xx",
      error: new ApiError(409, "CONFLICT", "already exists"),
      expected: "already exists",
    },
    {
      name: "maps network TypeError to network",
      error: new TypeError("Failed to fetch"),
      expected: "t:network",
    },
    {
      name: "falls back to generic for unknown values",
      error: "weird",
      expected: "t:generic",
    },
  ];

  for (const { name, error, expected } of cases) {
    it(name, () => {
      expect(getApiErrorMessage(error, t)).toBe(expected);
    });
  }
});
