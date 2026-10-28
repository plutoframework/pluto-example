import { renderHook } from "@testing-library/react";
import { expect, test } from "vitest";
import { useScopeGuard } from "./useScopeGuard";
test("late async UI effects cannot escape their original workspace", () => {
  const { result, rerender, unmount } = renderHook(
    ({ organization }) =>
      useScopeGuard({ actor: "alice", session: "session", organization }),
    { initialProps: { organization: "northstar" } },
  );
  const northstar = result.current;
  expect(northstar()).toBe(true);
  rerender({ organization: "harbor" });
  const harbor = result.current;
  expect(northstar()).toBe(false);
  expect(harbor()).toBe(true);
  unmount();
  expect(harbor()).toBe(false);
});
