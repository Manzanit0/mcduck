import { Signal, useComputed, useSignal } from "@preact/signals";
import { ReceiptStatus } from "../gen/receipts.v1/receipts_pb.ts";
import { JSX } from "preact/jsx-runtime";
import { SerializableReceipt } from "../lib/types.ts";
import { updateReceipt } from "../lib/receipts.ts";
import { Timestamp } from "@bufbuild/protobuf";

interface TableProps {
  receipts: SerializableReceipt[];
  url: string;
}

interface ViewReceipt extends SerializableReceipt {
  displayed: boolean;
  checked: boolean;
}

export default function SearcheableTable(props: TableProps) {
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

  return (
    <div class="sm:rounded-lg">
      <div class="flex flex-column sm:flex-row flex-wrap space-y-4 sm:space-y-0 items-center justify-between pb-4">
        <SearchBox onInput={filterReceipts} />
      </div>
      <table class="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
        <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
          <tr>
            <th scope="col" class="p-4">
              <Checkbox
                onInput={checkReceipts}
                checked={globallySelected.value}
              />
            </th>
            <th scope="col" class="px-6 py-3">
              Date
            </th>
            <th scope="col" class="px-6 py-3">
              Vendor
            </th>
            <th scope="col" class="px-6 py-3">
              Amount
            </th>
            <th scope="col" class="px-6 py-3">
              Status
            </th>
            <th scope="col" class="px-6 py-3">
              Action
            </th>
          </tr>
        </thead>
        <tbody>{displayedReceipts.value.map((r) => row(r, props.url))}</tbody>
      </table>
    </div>
  );
}

function row(r: Signal<ViewReceipt>, url: string) {
  const total = r.value.expenses.reduce((acc, ex) => (acc += ex.amount), 0n);

  const updateVendor = async (e: JSX.TargetedEvent<HTMLInputElement>) => {
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

  const updateDate = async (e: JSX.TargetedEvent<HTMLInputElement>) => {
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

  const updateStatus = async (status: string) => {
    if (status === r.value.status) {
      return;
    }

    r.value = { ...r.value, status: status };

    await updateReceipt(url, {
      id: r.peek().id,
      pendingReview: r.value.status === ReceiptStatus.PENDING_REVIEW.toString(),
    });

    console.log("updated status to", r.value.status);
  };

  // NOTE: the datepicker expects a date without the time. Since we
  // know" that they always come with time, we can just naively split.
  return (
    <tr class="bg-white border-b dark:bg-gray-800 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600">
      <td class="w-4 p-4">
        <Checkbox
          checked={r.value.checked}
          onInput={() => (r.value.checked = !r.value.checked)}
        />
      </td>
      <td class="px-6 py-4">
        <DatePicker value={r.value.date!.split("T")[0]} onChange={updateDate} />
      </td>
      <td class="px-6 py-4">
        <TextInput value={r.value.vendor} onfocusout={updateVendor} />
      </td>
      <td class="px-6 py-4">
        {formatEuro(total)}
      </td>
      <td class="px-6 py-4">
        <ReceiptStatusDropdown receipt={r} updateStatus={updateStatus} />
      </td>
      <td class="px-6 py-4">
        <a
          href="#"
          class="font-medium text-blue-600 dark:text-blue-500 hover:underline"
        >
          View
        </a>
      </td>
    </tr>
  );
}

function formatEuro(amount: bigint) {
  return new Intl.NumberFormat("de-DE", {
    style: "currency",
    currency: "EUR",
  }).format(Number(amount) / 100);
}

interface ReceiptStatusDropdownProps {
  receipt: Signal<ViewReceipt>;
  updateStatus: (status: string) => Promise<void>;
}

function ReceiptStatusDropdown(props: ReceiptStatusDropdownProps) {
  const open = useSignal(false);

  const dropdownOptions = useComputed(() => {
    const options = [pendingReview(false), reviewed(false)];
    switch (props.receipt.value.status) {
      case ReceiptStatus.PENDING_REVIEW.toString():
        options[0] = pendingReview(true);
        break;
      case ReceiptStatus.REVIEWED.toString():
        options[1] = reviewed(true);
        break;
      default:
        break;
    }

    return options;
  });

  const selectedDropdownOption = useComputed(() => {
    let option = na(false);
    switch (props.receipt.value.status) {
      case ReceiptStatus.PENDING_REVIEW.toString():
        option = pendingReview(false);
        break;
      case ReceiptStatus.REVIEWED.toString():
        option = reviewed(false);
        break;
      default:
        break;
    }
    return option;
  });

  return (
    <div>
      <div class="relative mt-2">
        <button
          type="button"
          class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 sm:text-sm sm:leading-6"
          aria-haspopup="listbox"
          aria-expanded="true"
          aria-labelledby="listbox-label"
          onClick={() => (open.value = !open.value)}
        >
          {selectedDropdownOption}
          <span class="pointer-events-none absolute inset-y-0 right-0 ml-3 flex items-center pr-2">
            <svg
              class="h-5 w-5 text-gray-400"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden="true"
            >
              <path
                fill-rule="evenodd"
                d="M10 3a.75.75 0 01.55.24l3.25 3.5a.75.75 0 11-1.1 1.02L10 4.852 7.3 7.76a.75.75 0 01-1.1-1.02l3.25-3.5A.75.75 0 0110 3zm-3.76 9.2a.75.75 0 011.06.04l2.7 2.908 2.7-2.908a.75.75 0 111.1 1.02l-3.25 3.5a.75.75 0 01-1.1 0l-3.25-3.5a.75.75 0 01.04-1.06z"
                clip-rule="evenodd"
              />
            </svg>
          </span>
        </button>

        <ul
          className={`absolute mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm ${
            open.value
              ? "z-10 opacity-100 transition ease-in duration-100"
              : "-z-10 opacity-0"
          }`}
          tabindex={-1}
          role="listbox"
          aria-labelledby="listbox-label"
          aria-activedescendant="listbox-option-3"
        >
          {dropdownOptions.value.map((x, index) => {
            const hovered = useSignal(false);
            return (
              <li
                class={`relative cursor-default select-none py-2 pl-3 pr-9 ${
                  hovered.value ? "bg-gray-100" : " text-gray-900"
                }`}
                role="option"
                onMouseEnter={() => (hovered.value = true)}
                onMouseLeave={() => (hovered.value = false)}
                onClick={async () => {
                  let status;
                  if (index === 0) {
                    status = ReceiptStatus.PENDING_REVIEW.toString();
                  } else {
                    status = ReceiptStatus.REVIEWED.toString();
                  }

                  // When the user selects and option, we can assume he wants the dropdown closed.
                  open.value = false;

                  await props.updateStatus(status);
                }}
              >
                {x}
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}

function checkmark() {
  return (
    <span class="absolute inset-y-0 right-0 flex items-center pr-4 text-grey-900">
      <svg
        class="h-5 w-5"
        viewBox="0 0 20 20"
        fill="currentColor"
        aria-hidden="true"
      >
        <path
          fill-rule="evenodd"
          d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z"
          clip-rule="evenodd"
        />
      </svg>
    </span>
  );
}

function reviewed(selected: boolean) {
  const fontClass = selected ? "font-semibold" : "font-normal";
  return (
    <>
      <div class="flex items-center">
        <div class="h-2.5 w-2.5 rounded-full bg-green-500 me-2"></div>
        <span class={`ml-3 block truncate ${fontClass}`}>
          Reviewed
        </span>
      </div>
      {selected ? checkmark() : <></>}
    </>
  );
}

function pendingReview(selected: boolean) {
  const fontClass = selected ? "font-semibold" : "font-normal";
  return (
    <>
      <div class="flex items-center">
        <div class="h-2.5 w-2.5 rounded-full bg-red-500 me-2"></div>
        <span class={`ml-3 block truncate ${fontClass}`}>
          Pending Review
        </span>
      </div>
      {selected ? checkmark() : <></>}
    </>
  );
}

function na(selected: boolean) {
  const fontClass = selected ? "font-semibold" : "font-normal";
  return (
    <>
      <div class="flex items-center">
        <div class="h-2.5 w-2.5 rounded-full bg-yellow-500 me-2"></div>
        <span class={`ml-3 block truncate ${fontClass}`}>
          N/a
        </span>
      </div>
      {selected ? checkmark() : <></>}
    </>
  );
}

interface TextInputProps {
  value: string;
  onfocusout: (e: JSX.TargetedEvent<HTMLInputElement>) => Promise<void>;
}

function TextInput(props: TextInputProps) {
  return (
    <div>
      <div class="relative mt-2 rounded-md shadow-sm">
        <input
          type="text"
          class="block w-full rounded-md border-0 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm sm:leading-6 focus:outline-none focus:ring-2 focus:ring-gray-500"
          placeholder="0.00"
          value={props.value}
          onfocusout={props.onfocusout}
        />
      </div>
    </div>
  );
}

interface DatepickerProps {
  value: string;
  onChange: (e: JSX.TargetedEvent<HTMLInputElement>) => Promise<void>;
}

function DatePicker(props: DatepickerProps) {
  return (
    <div>
      <div class="relative mt-2 rounded-md shadow-sm">
        <input
          type="date"
          class="block w-full rounded-md border-0 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm sm:leading-6 focus:outline-none focus:ring-2 focus:ring-gray-500"
          placeholder="0.00"
          value={props.value}
          onChange={props.onChange}
        />
      </div>
    </div>
  );
}

interface SearchBoxProps {
  onInput: (e: JSX.TargetedEvent<HTMLInputElement>) => void;
}

function SearchBox({ onInput }: SearchBoxProps) {
  return (
    <>
      <label for="table-search" class="sr-only">
        Search
      </label>
      <div class="relative">
        <div class="absolute inset-y-0 left-0 rtl:inset-r-0 rtl:right-0 flex items-center ps-3 pointer-events-none">
          <svg
            class="w-5 h-5 text-gray-500 dark:text-gray-400"
            aria-hidden="true"
            fill="currentColor"
            viewBox="0 0 20 20"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"
            >
            </path>
          </svg>
        </div>
        <input
          onInput={onInput}
          type="text"
          id="table-search"
          class="block p-2 ps-10 text-sm text-gray-900 border border-gray-300 rounded-lg w-80 bg-gray-50 focus:ring-gray-500 focus:border-gray-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-gray-500 dark:focus:border-gray-500"
          placeholder="Search for items"
        />
      </div>
    </>
  );
}

interface CheckboxProps {
  checked: boolean;
  onInput: () => void;
}

function Checkbox({ checked, onInput }: CheckboxProps) {
  return (
    <div class="flex items-center">
      <input
        id="checkbox-all-search"
        type="checkbox"
        class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 dark:focus:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
        onInput={onInput}
        checked={checked}
      />
      <label for="checkbox-all-search" class="sr-only">
        checkbox
      </label>
    </div>
  );
}
