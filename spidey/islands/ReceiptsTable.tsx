import GenericTable from "../components/GenericTable.tsx";
import { Signal, useComputed, useSignal } from "@preact/signals";
import { ReceiptStatus } from "../gen/receipts.v1/receipts_pb.ts";
import { JSX } from "preact/jsx-runtime";
import { SerializableReceipt } from "../lib/types.ts";
import { updateReceipt } from "../lib/receipts.ts";
import { Timestamp } from "@bufbuild/protobuf";
import TextInput from "../components/TextInput.tsx";
import Checkbox from "../components/Checkbox.tsx";
import SearchBox from "../components/SearchBox.tsx";
import DatePicker from "../components/DatePicker.tsx";
import ReceiptStatusDropdown from "./ReceiptStatusDropdown.tsx";
import FormattedMoney from "../components/FormattedMoney.tsx";

interface TableProps {
  receipts: SerializableReceipt[];
  url: string;
}

interface ViewReceipt extends SerializableReceipt {
  displayed: boolean;
  checked: boolean;
}

export default function ReceiptsTable(props: TableProps) {
  const mapped = props.receipts.map((x) => {
    return useSignal({
      ...x,
      checked: false,
      displayed: true,
    });
  });

  const globallySelected = useSignal(false);
  const searchText = useSignal("");
  const allReceipts = useSignal(mapped);
  const displayedReceipts = useComputed(() =>
    mapped.filter((x) => {
      return x.value.vendor
        .toLowerCase()
        .includes(searchText.value.toLowerCase());
    })
  );

  const filterReceipts = (e: JSX.TargetedEvent<HTMLInputElement>) => {
    searchText.value = e.currentTarget.value;

    // Set the global checkbox depending on if all the rows are checked or not.
    const checked = displayedReceipts.value.filter((x) => x.peek().checked);
    globallySelected.value = checked.length === displayedReceipts.value.length;
  };

  const checkReceipts = () => {
    globallySelected.value = !globallySelected.value;

    for (const r of allReceipts.value) {
      for (const d of displayedReceipts.value) {
        if (r.value.id === d.value.id) {
          r.value.checked = globallySelected.value;
          break;
        }
      }
    }
  };

  const updateVendor = async (
    e: JSX.TargetedEvent<HTMLInputElement>,
    r: Signal<ViewReceipt>,
  ) => {
    if (!e.currentTarget || e.currentTarget.value === "") {
      return;
    }

    const vendor = e.currentTarget.value;
    if (vendor === r.value.vendor) {
      return;
    }

    r.value = { ...r.value, vendor: vendor };

    await updateReceipt(props.url, { id: r.peek().id, vendor: vendor });
    console.log("updated vendor to", vendor);
  };

  const updateDate = async (
    e: JSX.TargetedEvent<HTMLInputElement>,
    r: Signal<ViewReceipt>,
  ) => {
    if (!e.currentTarget || e.currentTarget.value === "") {
      return;
    }

    const date = e.currentTarget.value;
    if (date === r.value.date) {
      return;
    }

    r.value = { ...r.value, date: date };

    await updateReceipt(props.url, {
      id: r.peek().id,
      date: Timestamp.fromDate(new Date(date)),
    });
    console.log("updated date to", date);
  };

  const updateStatus = async (status: string, r: Signal<ViewReceipt>) => {
    if (status === r.value.status) {
      return;
    }

    r.value = { ...r.value, status: status };

    await updateReceipt(props.url, {
      id: r.peek().id,
      pendingReview: r.value.status === ReceiptStatus.PENDING_REVIEW.toString(),
    });

    console.log("updated status to", r.value.status);
  };

  return (
    <div class="sm:rounded-lg">
      <div class="flex flex-column sm:flex-row flex-wrap space-y-4 sm:space-y-0 items-center justify-between pb-4">
        <SearchBox onInput={filterReceipts} />
      </div>
      <GenericTable
        data={displayedReceipts.value}
        columns={[
          {
            header: (
              <Checkbox
                onInput={checkReceipts}
                checked={globallySelected.value}
              />
            ),
            accessor: (r) => (
              <Checkbox
                checked={r.value.checked}
                onInput={() => (r.value.checked = !r.value.checked)}
              />
            ),
          },
          {
            header: <span>Date</span>,
            accessor: (r) => (
              <DatePicker
                value={r.value.date!.split("T")[0]}
                onChange={(e) => updateDate(e, r)}
              />
            ),
          },
          {
            header: <span>Vendor</span>,
            accessor: (r) => (
              <TextInput
                value={r.value.vendor}
                onfocusout={(e) => updateVendor(e, r)}
              />
            ),
          },
          {
            header: <span>Amount</span>,
            accessor: (r) => (
              <FormattedMoney
                currency="EUR"
                amount={Number(
                  r.value.expenses.reduce((acc, ex) => (acc += ex.amount), 0n),
                )}
              />
            ),
          },
          {
            header: <span>Status</span>,
            accessor: (r) => (
              <ReceiptStatusDropdown
                receipt={r}
                updateStatus={(status) => updateStatus(status, r)}
              />
            ),
          },
          {
            header: <span>Action</span>,
            accessor: (r) => (
              <a
                href={`receipts/${r.value.id}`}
                class="font-medium text-blue-600 dark:text-blue-500 hover:underline"
              >
                View
              </a>
            ),
          },
        ]}
      />
    </div>
  );
}
