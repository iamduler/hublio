import { describe, expect, it } from "vitest";
import { unwrapData } from "./types";

describe("unwrapData", () => {
  const cases: Array<{
    name: string;
    input: unknown;
    expected: unknown;
  }> = [
    {
      name: "unwraps an envelope with a data field",
      input: { status: "success", data: { id: "1" } },
      expected: { id: "1" },
    },
    {
      name: "unwraps an array data field",
      input: { status: "success", data: [1, 2, 3] },
      expected: [1, 2, 3],
    },
    {
      name: "returns the body when there is no data field",
      input: { id: "2" },
      expected: { id: "2" },
    },
    {
      name: "returns the body when data is undefined",
      input: { status: "success", data: undefined },
      expected: { status: "success", data: undefined },
    },
  ];

  for (const { name, input, expected } of cases) {
    it(name, () => {
      expect(unwrapData(input)).toEqual(expected);
    });
  }
});
