import { getRequestConfig } from "next-intl/server";
import messages from "../messages/en.json";

// Admin is single-locale (en) for now. This satisfies next-intl's server
// config requirement; locale routing can be added later if needed.
export default getRequestConfig(async () => ({
  locale: "en",
  messages,
}));
