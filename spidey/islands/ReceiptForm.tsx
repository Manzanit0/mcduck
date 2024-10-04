import { Signal, useSignal } from "@preact/signals";
import DatePicker from "../components/DatePicker.tsx";
import TextInput from "../components/TextInput.tsx";
import Label from "../components/Label.tsx";
import ReceiptStatusDropdown from "./ReceiptStatusDropdown.tsx";
import { SerializableReceipt } from "../lib/types.ts";
import { JSX } from "preact/jsx-runtime";
import { Timestamp } from "@bufbuild/protobuf";
import { ReceiptStatus } from "../gen/receipts.v1/receipts_pb.ts";
import { updateReceipt } from "../lib/receipts.ts";

interface ReceiptFormProps {
  receipt: SerializableReceipt;
  url: string;
}

export default function ReceiptForm({ receipt, url }: ReceiptFormProps) {
  const r = useSignal(receipt);

  const updateVendor = async (
    e: JSX.TargetedEvent<HTMLInputElement>,
    r: Signal<SerializableReceipt>
  ) => {
    if (!e.currentTarget || e.currentTarget.value === "") {
      return;
    }

    const vendor = e.currentTarget.value;
    if (vendor === r.value.vendor) {
      return;
    }

    r.value = { ...r.value, vendor: vendor };

    await updateReceipt(url, { id: r.peek().id, vendor: vendor });
    console.log("updated vendor to", vendor);
  };

  const updateDate = async (
    e: JSX.TargetedEvent<HTMLInputElement>,
    r: Signal<SerializableReceipt>
  ) => {
    if (!e.currentTarget || e.currentTarget.value === "") {
      return;
    }

    const date = e.currentTarget.value;
    if (date === r.value.date) {
      return;
    }

    r.value = { ...r.value, date: date };

    await updateReceipt(url, {
      id: r.peek().id,
      date: Timestamp.fromDate(new Date(date)),
    });
    console.log("updated date to", date);
  };

  const updateStatus = async (
    status: number,
    r: Signal<SerializableReceipt>
  ) => {
    if (status === r.value.status) {
      return;
    }

    r.value = { ...r.value, status: status };

    await updateReceipt(url, {
      id: r.peek().id,
      pendingReview: r.value.status === ReceiptStatus.PENDING_REVIEW,
    });

    console.log("updated status to", r.value.status);
  };

  return (
    <div>
      <h2 class="text-base font-semibold leading-7 text-gray-900">
        Receipt Information
      </h2>
      <p class="mt-1 text-sm leading-6 text-gray-600">
        Updating the date of the receipt will update the date of all the
        expenses.
      </p>
      <div class="mt-10 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-3">
        <div class="mt-2 col-span-1">
          <Label>Vendor</Label>
          <TextInput
            value={r.value.vendor}
            onfocusout={(e) => updateVendor(e, r)}
          />
        </div>
        <div class="mt-2 col-span-1">
          <Label>Status</Label>
          <ReceiptStatusDropdown
            status={r.value.status}
            updateStatus={(status) => updateStatus(status, r)}
          />
        </div>
        <div class="mt-2 col-start-1 cols-end-2">
          <Label>Date</Label>
          <DatePicker
            value={r.value.date!.split("T")[0]}
            onChange={(e) => updateDate(e, r)}
          />
        </div>
      </div>
    </div>
  );
}
