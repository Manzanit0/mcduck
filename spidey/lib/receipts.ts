import { createConnectTransport } from "@connectrpc/connect-web";
import { createPromiseClient } from "@connectrpc/connect";
import { ReceiptsService } from "../gen/receipts.v1/receipts_connect.ts";
import { UpdateReceiptRequest } from "../gen/receipts.v1/receipts_pb.ts";
import { getAuthTokenFromBrowser } from "./auth.ts";
import { PartialMessage } from "@bufbuild/protobuf";

export function updateReceipt(host: string, body: PartialMessage<UpdateReceiptRequest>) {
  const client = createPromiseClient(
    ReceiptsService,
    createConnectTransport({
      baseUrl: host,
    }),
  );

  const { authToken } = getAuthTokenFromBrowser();

  return client.updateReceipt(body, {
    headers: { authorization: `Bearer ${authToken}` },
  });
}
