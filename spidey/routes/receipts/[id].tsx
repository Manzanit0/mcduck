import { RouteContext } from "$fresh/server.ts";
// import * as base64 from "$std/encoding/base64";
import * as base64 from "jsr:@std/encoding/base64";
import ExpensesTable from "../../islands/ExpensesTable.tsx";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ReceiptsService } from "../../gen/receipts.v1/receipts_connect.ts";
import {
  mapExpensesToSerializable,
  mapReceiptsToSerializable,
} from "../../lib/types.ts";

import { AuthState } from "../../lib/auth.ts";
import ReceiptForm from "../../islands/ReceiptForm.tsx";

const url = Deno.env.get("API_HOST")!;

export default async function Single(_: Request, ctx: RouteContext<AuthState>) {
  console.log("get receipt");
  const transport = createConnectTransport({
    baseUrl: url!,
  });

  const client = createPromiseClient(ReceiptsService, transport);

  const res = await client.getReceipt(
    { id: BigInt(ctx.params.id) },
    {
      headers: { authorization: `Bearer ${ctx.state.authToken}` },
    }
  );

  const receipt = mapReceiptsToSerializable([res.receipt!])[0];
  const encoded = base64.encodeBase64(res.receipt!.file);

  return (
    <div class="m-6">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
        Receipt #{ctx.params.id}
      </h2>
      <div class="mt-10 grid grid-cols-3 gap-4 items-center">
        <div class="col-span-1">
          <img
            class="object-contain"
            src={`data:image/png;base64, ${encoded}`}
            alt="Receipt image"
          />
        </div>
        <div class="col-span-2 p-5">
          <ReceiptForm receipt={receipt} url={url} />
          <div class="mt-10">
            <h2 class="text-base font-semibold leading-7 text-gray-900">
              Expenses
            </h2>
            <div class="mt-2">
              <ExpensesTable
                expenses={mapExpensesToSerializable(res.receipt!.expenses)}
                url={url}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
