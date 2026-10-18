import React from "react";
import { fireEvent, render, screen } from "@testing-library/react-native";
import { router } from "expo-router";
import PairRoute from "../../../app/pair";
import { setClock } from "../../clock/clock";
import PairScreen, { timeLeft } from "../PairScreen";
import { usePairing, type PairingView } from "../usePairing";

jest.mock("expo-router", () => ({ router: { replace: jest.fn() } }));
jest.mock("../usePairing", () => ({ usePairing: jest.fn() }));

const at = (time: string) => new Date(`2026-10-17T${time}`);

describe("PairScreen", () => {
  beforeEach(() => setClock(at("10:00:00")));
  afterEach(() => {
    setClock(null);
    jest.clearAllMocks();
  });

  it("shows the code, where to enter it and how long it lasts", async () => {
    await render(<PairScreen view={{ kind: "showing", code: "K7M4QP", expiresAt: at("10:06:42") }} onNewCode={() => {}} />);

    expect(screen.getByText("Set up this tablet")).toBeTruthy();
    expect(screen.getByText("Enter this code in the family app")).toBeTruthy();
    expect(screen.getByTestId("pairing-code")).toHaveTextContent("K7M4QP");
    expect(screen.getByText("Expires in 6:42")).toBeTruthy();
  });

  it("asks for a new code when tapped", async () => {
    const onNewCode = jest.fn();
    await render(<PairScreen view={{ kind: "showing", code: "K7M4QP", expiresAt: at("10:10:00") }} onNewCode={onNewCode} />);

    fireEvent.press(screen.getByRole("button", { name: "New code" }));

    expect(onNewCode).toHaveBeenCalled();
  });

  it.each<[PairingView, string]>([
    [{ kind: "starting" }, "Getting a code…"],
    [{ kind: "offline" }, "Can't reach the server. Trying again…"],
  ])("says what it's doing while there's no code (%o)", async (view, text) => {
    await render(<PairScreen view={view} onNewCode={() => {}} />);

    expect(screen.getByText(text)).toBeTruthy();
    expect(screen.queryByRole("button")).toBeNull();
  });

  it.each([
    ["10:06:42", "6:42"],
    ["10:00:00.400", "0:01"],
    ["10:00:00", "0:00"],
    ["09:59:00", "0:00"],
  ])("at 10:00, a code expiring at %s has %s left", (expires, shown) => {
    expect(timeLeft(at(expires), at("10:00:00"))).toBe(shown);
  });
});

describe("PairRoute", () => {
  afterEach(() => jest.clearAllMocks());

  it("goes home once paired", async () => {
    jest.mocked(usePairing).mockReturnValue({ view: { kind: "paired" }, newCode: jest.fn() });

    await render(<PairRoute />);

    expect(router.replace).toHaveBeenCalledWith("/");
  });

  it("stays while waiting to be paired", async () => {
    jest.mocked(usePairing).mockReturnValue({ view: { kind: "starting" }, newCode: jest.fn() });

    await render(<PairRoute />);

    expect(router.replace).not.toHaveBeenCalled();
  });
});
