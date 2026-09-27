import { useState, useEffect } from 'react';
import type { ConfigurationWithMetadata } from './AgendaList';
import { useAuth } from '../context/AuthContext';

const API_URL = import.meta.env.VITE_API_URL;

type OccurrenceType = 'daily' | 'weekly' | 'monthly' | 'yearly' | 'single';

const DAYS_OF_WEEK = ['Sun', 'Mon', 'Tues', 'Wed', 'Thur', 'Fri', 'Sat'] as const;
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'] as const;

interface SelectedModule {
    moduleId: number;
    moduleType: string;
    label: string;
}

interface ConfigurationFormProps {
    /** When provided the form runs in edit mode, pre-filled with this config */
    existingConfig?: ConfigurationWithMetadata;
    /** When provided the form runs in create mode for this module */
    selectedModule?: SelectedModule;
    /** Current number of items in the agenda (used for default position on create) */
    currentItemCount: number;
    /** Callback after a successful save / update / delete so they can refresh */
    onSave: () => void;
    onCancel: () => void;
    /** The currently viewed date (YYYY-MM-DD) — used for creating single-day exceptions */
    currentDate?: string;
}

export default function ConfigurationForm({
    existingConfig,
    selectedModule,
    currentItemCount,
    onSave,
    onCancel,
    currentDate,
}: ConfigurationFormProps) {
    const { user, activeCarePartnerId } = useAuth();
    const isEditMode = Boolean(existingConfig);
    const [occurrenceType, setOccurrenceType] = useState<OccurrenceType>('daily');
    const [weeklyDays, setWeeklyDays] = useState<string[]>([]);
    const [monthlyDays, setMonthlyDays] = useState<string[]>([]);
    const [yearlyDates, setYearlyDates] = useState<string[]>([]);
    const [yearlyMonth, setYearlyMonth] = useState<string>('Jan');
    const [yearlyDay, setYearlyDay] = useState<number>(1);
    const [singleDate, setSingleDate] = useState<string>('');
    const [duration, setDuration] = useState<number>(30);
    const [position, setPosition] = useState<number>(currentItemCount + 1);
    const [timed, setTimed] = useState<boolean>(false);
    const [time, setTime] = useState<string>('');
    const [status, setStatus] = useState('');
    const [saving, setSaving] = useState(false);
    const [deleteModal, setDeleteModal] = useState<'closed' | 'choose' | 'confirm-single'>('closed');

    // Pre-fill for edit mode
    useEffect(() => {
        if (!existingConfig) return;
        setOccurrenceType(existingConfig.occurence_type as OccurrenceType);
        setDuration(existingConfig.duration);
        setPosition(existingConfig.position);
        setTimed(existingConfig.timed ?? false);
        setTime(existingConfig.time ?? '');

        const def = existingConfig.occurence_definition || [];
        switch (existingConfig.occurence_type) {
            case 'weekly':
                setWeeklyDays(def);
                break;
            case 'monthly':
                setMonthlyDays(def);
                break;
            case 'yearly':
                setYearlyDates(def);
                break;
            case 'single':
                setSingleDate(def[0] || '');
                break;
        }
    }, [existingConfig]);

    // Build occurence_definition from UI state
    const buildOccurrenceDefinition = (): string[] => {
        switch (occurrenceType) {
            case 'daily':
                return [];
            case 'weekly':
                return weeklyDays;
            case 'monthly':
                return monthlyDays;
            case 'yearly':
                return yearlyDates;
            case 'single':
                return singleDate ? [singleDate] : [];
            default:
                return [];
        }
    };

    // Toggle a day in the weekly selection
    const toggleWeeklyDay = (day: string) => {
        setWeeklyDays((prev) =>
            prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day]
        );
    };

    // Toggle a day number in the monthly selection
    const toggleMonthlyDay = (day: string) => {
        setMonthlyDays((prev) =>
            prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day]
        );
    };

    // Add a yearly date entry
    const addYearlyDate = () => {
        // Validate day against the selected month
        const monthIndex = MONTHS.indexOf(yearlyMonth as typeof MONTHS[number]);
        // Use a non-leap year (e.g. 2023) to get max days; Feb max = 29 to allow leap-year dates
        const maxDay = monthIndex === 1 ? 29 : new Date(2023, monthIndex + 1, 0).getDate();
        if (yearlyDay < 1 || yearlyDay > maxDay) {
            setStatus(`Error: ${yearlyMonth} does not have day ${yearlyDay}.`);
            return;
        }
        const entry = `${yearlyMonth} ${yearlyDay}`;
        if (!yearlyDates.includes(entry)) {
            setYearlyDates((prev) => [...prev, entry]);
        }
    };

    // Remove a yearly date entry
    const removeYearlyDate = (entry: string) => {
        setYearlyDates((prev) => prev.filter((d) => d !== entry));
    };

    // Derive the display label
    const moduleLabel = isEditMode
        ? getModuleLabel(existingConfig!)
        : selectedModule?.label || 'Unknown Module';
    
    const moduleType = isEditMode
        ? existingConfig?.module_type
        : selectedModule?.moduleType;

    const isMediaType = moduleType === 'video' || moduleType === 'audio';

    useEffect(() => {
        if (isMediaType) {
            setDuration(0);
        }
    }, [isMediaType]);

    const handleCreate = async () => {
        if (!selectedModule) return;
        if (!activeCarePartnerId || !user?.id) {
            setStatus('Error: no carepartner is linked to this caregiver account.');
            return;
        }

        // Validate occurrence definitions are non-empty for types that require them
        if (occurrenceType === 'weekly' && weeklyDays.length === 0) {
            setStatus('Error: please select at least one day of the week.');
            return;
        }
        if (occurrenceType === 'monthly' && monthlyDays.length === 0) {
            setStatus('Error: please select at least one day of the month.');
            return;
        }
        if (occurrenceType === 'yearly' && yearlyDates.length === 0) {
            setStatus('Error: please add at least one yearly date.');
            return;
        }
        if (occurrenceType === 'single' && !singleDate) {
            setStatus('Error: please select a date.');
            return;
        }

        setSaving(true);
        setStatus('');

        try {
            const configEntry = {
                id: null,
                module_id: selectedModule.moduleId,
                occurence_type: occurrenceType,
                occurence_definition: buildOccurrenceDefinition(),
                position: position,
                duration: duration,
                timed: timed,
                time: timed && time ? time + ':00' : null,
            };

            const res = await fetch(`${API_URL}/configurations/insert`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    carepartner_id: activeCarePartnerId,
                    json_config: JSON.stringify([configEntry]),
                }),
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed to create configuration');
            }

            setStatus('Configuration created successfully!');
            onSave();
        } catch (err) {
            setStatus(`Error: ${err instanceof Error ? err.message : String(err)}`);
        } finally {
            setSaving(false);
        }
    };

    const handleUpdate = async () => {
        if (!existingConfig) return;

        // Validate occurrence definitions are non-empty for types that require them
        if (occurrenceType === 'weekly' && weeklyDays.length === 0) {
            setStatus('Error: please select at least one day of the week.');
            return;
        }
        if (occurrenceType === 'monthly' && monthlyDays.length === 0) {
            setStatus('Error: please select at least one day of the month.');
            return;
        }
        if (occurrenceType === 'yearly' && yearlyDates.length === 0) {
            setStatus('Error: please add at least one yearly date.');
            return;
        }
        if (occurrenceType === 'single' && !singleDate) {
            setStatus('Error: please select a date.');
            return;
        }

        setSaving(true);
        setStatus('');

        try {
            const configEntry = {
                id: existingConfig.id,
                module_id: existingConfig.module_id,
                occurence_type: occurrenceType,
                occurence_definition: buildOccurrenceDefinition(),
                position: position,
                duration: duration,
                timed: timed,
                time: timed && time ? time + ':00' : null,
            };

            const res = await fetch(`${API_URL}/configurations/insert`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    carepartner_id: activeCarePartnerId,
                    json_config: JSON.stringify([configEntry]),
                }),
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed to update configuration');
            }

            setStatus('Configuration updated successfully!');
            onSave();
        } catch (err) {
            setStatus(`Error: ${err instanceof Error ? err.message : String(err)}`);
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = () => {
        if (!existingConfig) return;

        // For non-single occurrence types, show the choice modal
        if (existingConfig.occurence_type !== 'single' && currentDate) {
            setDeleteModal('choose');
        } else {
            // For single-day items, show a simple confirm modal
            setDeleteModal('confirm-single');
        }
    };

    const handleSkipThisDay = async () => {
        if (!existingConfig || !currentDate) return;
        setDeleteModal('closed');
        setSaving(true);
        setStatus('');
        try {
            const res = await fetch(`${API_URL}/exceptions`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    configuration_id: existingConfig.id,
                    exception_type: 'single',
                    exception_definition: [currentDate],
                }),
            });
            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed to create exception');
            }
            setStatus('This day has been skipped.');
            onSave();
        } catch (err) {
            setStatus(`Error: ${err instanceof Error ? err.message : String(err)}`);
        } finally {
            setSaving(false);
        }
    };

    const handleDeleteAll = async () => {
        if (!existingConfig) return;
        setDeleteModal('closed');
        setSaving(true);
        setStatus('');

        try {
            const res = await fetch(`${API_URL}/configurations/delete`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    carepartner_id: activeCarePartnerId,
                    id_list: [existingConfig.id],
                }),
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed to delete configuration');
            }

            setStatus('Configuration removed successfully!');
            onSave();
        } catch (err) {
            setStatus(`Error: ${err instanceof Error ? err.message : String(err)}`);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="w-full flex flex-col items-center px-6 py-4">
            <h1 className="text-3xl font-semibold mb-6">
                {isEditMode ? 'Edit Agenda Item' : 'Add to Agenda'}
            </h1>

            <div className="border p-8 max-w-3xl w-full rounded-lg bg-white shadow space-y-6">
                {/* Module Info (read-only) */}
                <div>
                    <label className="text-lg font-medium text-gray-700">Module</label>
                    <div className="w-full mt-2 p-3 border rounded bg-gray-100 text-gray-700">
                        {moduleLabel}
                    </div>
                </div>

                {/* Occurrence Type */}
                <div>
                    <label className="text-lg font-medium text-gray-700">Recurrence</label>
                    <select
                        value={occurrenceType}
                        onChange={(e) => setOccurrenceType(e.target.value as OccurrenceType)}
                        className="w-full mt-2 p-3 border rounded bg-purple-50"
                    >
                        <option value="daily">Daily</option>
                        <option value="weekly">Weekly</option>
                        <option value="monthly">Monthly</option>
                        <option value="yearly">Yearly</option>
                        <option value="single">Single Day</option>
                    </select>
                </div>

                {/* Occurrence Definition — varies by type */}
                {occurrenceType === 'single' && (
                    <div>
                        <label className="text-lg font-medium text-gray-700">Date</label>
                        <input
                            type="date"
                            value={singleDate}
                            onChange={(e) => setSingleDate(e.target.value)}
                            className="w-full mt-2 p-3 border rounded bg-purple-50"
                        />
                    </div>
                )}

                {occurrenceType === 'weekly' && (
                    <div>
                        <label className="text-lg font-medium text-gray-700">Days of the Week</label>
                        <div className="flex flex-wrap gap-2 mt-2">
                            {DAYS_OF_WEEK.map((day) => (
                                <button
                                    key={day}
                                    type="button"
                                    onClick={() => toggleWeeklyDay(day)}
                                    className={`px-4 py-2 rounded-full border text-sm font-medium transition ${
                                        weeklyDays.includes(day)
                                            ? 'bg-teal-300 border-teal-400 text-gray-800'
                                            : 'bg-white border-gray-300 text-gray-600 hover:bg-gray-100'
                                    }`}
                                >
                                    {day}
                                </button>
                            ))}
                        </div>
                    </div>
                )}

                {occurrenceType === 'monthly' && (
                    <div>
                        <label className="text-lg font-medium text-gray-700">Days of the Month</label>
                        <div className="grid grid-cols-7 gap-2 mt-2">
                            {Array.from({ length: 31 }, (_, i) => String(i + 1)).map((day) => (
                                <button
                                    key={day}
                                    type="button"
                                    onClick={() => toggleMonthlyDay(day)}
                                    className={`px-2 py-2 rounded border text-sm font-medium transition ${
                                        monthlyDays.includes(day)
                                            ? 'bg-teal-300 border-teal-400 text-gray-800'
                                            : 'bg-white border-gray-300 text-gray-600 hover:bg-gray-100'
                                    }`}
                                >
                                    {day}
                                </button>
                            ))}
                        </div>
                    </div>
                )}

                {occurrenceType === 'yearly' && (
                    <div>
                        <label className="text-lg font-medium text-gray-700">Dates (Month &amp; Day)</label>
                        <div className="flex gap-2 mt-2">
                            <select
                                value={yearlyMonth}
                                onChange={(e) => setYearlyMonth(e.target.value)}
                                className="p-3 border rounded bg-purple-50"
                            >
                                {MONTHS.map((m) => (
                                    <option key={m} value={m}>
                                        {m}
                                    </option>
                                ))}
                            </select>
                            <input
                                type="number"
                                min={1}
                                max={31}
                                value={yearlyDay}
                                onChange={(e) => setYearlyDay(Number(e.target.value))}
                                className="w-20 p-3 border rounded bg-purple-50"
                            />
                            <button
                                type="button"
                                onClick={addYearlyDate}
                                className="px-4 py-2 bg-teal-200 hover:bg-teal-300 rounded font-medium"
                            >
                                Add
                            </button>
                        </div>
                        {yearlyDates.length > 0 && (
                            <div className="flex flex-wrap gap-2 mt-3">
                                {yearlyDates.map((entry) => (
                                    <span
                                        key={entry}
                                        className="flex items-center gap-1 px-3 py-1 bg-teal-100 border border-teal-300 rounded-full text-sm"
                                    >
                                        {entry}
                                        <button
                                            type="button"
                                            onClick={() => removeYearlyDate(entry)}
                                            className="text-red-500 hover:text-red-700 font-bold"
                                        >
                                            ×
                                        </button>
                                    </span>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {/* Timed Toggle */}
                <div>
                    <label className="flex items-center gap-3 cursor-pointer">
                        <span className="text-lg font-medium text-gray-700">Scheduled Time</span>
                        <div
                            onClick={() => {
                                setTimed(!timed);
                                if (timed) setTime('');
                            }}
                            className={`relative w-12 h-6 rounded-full transition-colors ${
                                timed ? 'bg-teal-400' : 'bg-gray-300'
                            }`}
                        >
                            <div
                                className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform ${
                                    timed ? 'translate-x-6' : ''
                                }`}
                            />
                        </div>
                    </label>
                    {timed && (
                        <input
                            type="time"
                            value={time}
                            onChange={(e) => setTime(e.target.value)}
                            className="w-full mt-2 p-3 border rounded bg-purple-50"
                        />
                    )}
                </div>

                {/* Duration */}
                <div>
                    <label className="text-lg font-medium text-gray-700">Duration (seconds)</label>
                    <input
                        type="number"
                        min={1}
                        value={duration}
                        onChange={(e) => setDuration(Number(e.target.value))}
                        disabled={isMediaType}
                        className={`w-full mt-2 p-3 border rounded ${
                            isMediaType
                                ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                : 'bg-purple-50'
                        }`}
                    />
                    {isMediaType && (
                        <p className="text-sm text-gray-500 mt-1">
                            Duration is automatically determined for video/audio.
                        </p>
                    )}
                </div>

                {/* Action Buttons */}
                {isEditMode ? (
                    <div className="flex flex-col gap-3">
                        <button
                            disabled={saving}
                            onClick={handleUpdate}
                            className="w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded disabled:opacity-50"
                        >
                            {saving ? 'Saving…' : 'Update'}
                        </button>
                        <button
                            disabled={saving}
                            onClick={handleDelete}
                            className="w-full bg-red-100 hover:bg-red-200 text-red-700 font-semibold py-3 rounded disabled:opacity-50"
                        >
                            {saving ? 'Deleting…' : 'Remove from Agenda'}
                        </button>
                        <button
                            disabled={saving}
                            onClick={onCancel}
                            className="w-full border border-gray-300 text-gray-600 font-semibold py-3 rounded hover:bg-gray-50"
                        >
                            Cancel
                        </button>
                    </div>
                ) : (
                    <div className="flex flex-col gap-3">
                        <button
                            disabled={saving}
                            onClick={handleCreate}
                            className="w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded disabled:opacity-50"
                        >
                            {saving ? 'Saving…' : 'Add to Agenda'}
                        </button>
                        <button
                            disabled={saving}
                            onClick={onCancel}
                            className="w-full border border-gray-300 text-gray-600 font-semibold py-3 rounded hover:bg-gray-50"
                        >
                            Cancel
                        </button>
                    </div>
                )}

                {status && (
                    <p
                        className={`mt-2 text-sm text-center ${
                            status.startsWith('Error') ? 'text-red-600' : 'text-green-600'
                        }`}
                    >
                        {status}
                    </p>
                )}
            </div>

            {/* Delete Confirmation Modal */}
            {deleteModal !== 'closed' && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-lg shadow-lg max-w-md w-full mx-4 p-8">
                        {deleteModal === 'choose' ? (
                            <>
                                <h2 className="text-2xl font-semibold text-gray-800 text-center mb-2">Remove Item</h2>
                                <p className="text-gray-600 text-center mb-6">
                                    Would you like to skip this item just for today, or remove it from the agenda entirely?
                                </p>
                                <div className="flex flex-col gap-3">
                                    <button
                                        onClick={handleSkipThisDay}
                                        className="w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                                    >
                                        Skip Just This Day
                                    </button>
                                    <button
                                        onClick={handleDeleteAll}
                                        className="w-full bg-red-100 hover:bg-red-200 text-red-700 font-semibold py-3 rounded"
                                    >
                                        Delete Entire Occurrence
                                    </button>
                                    <button
                                        onClick={() => setDeleteModal('closed')}
                                        className="w-full border border-gray-300 text-gray-600 font-semibold py-3 rounded hover:bg-gray-50"
                                    >
                                        Cancel
                                    </button>
                                </div>
                            </>
                        ) : (
                            <>
                                <h2 className="text-2xl font-semibold text-gray-800 text-center mb-2">Confirm Removal</h2>
                                <p className="text-gray-600 text-center mb-6">
                                    Are you sure you want to remove this item from the agenda?
                                </p>
                                <div className="flex flex-col gap-3">
                                    <button
                                        onClick={handleDeleteAll}
                                        className="w-full bg-red-100 hover:bg-red-200 text-red-700 font-semibold py-3 rounded"
                                    >
                                        Remove
                                    </button>
                                    <button
                                        onClick={() => setDeleteModal('closed')}
                                        className="w-full border border-gray-300 text-gray-600 font-semibold py-3 rounded hover:bg-gray-50"
                                    >
                                        Cancel
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

function getModuleLabel(cfg: ConfigurationWithMetadata): string {
    const meta = cfg.metadata;
    if (!meta) return `Module ${cfg.module_id}`;

    switch (cfg.module_type) {
        case 'video':
            return `Video - ${(meta as any).displayed_name ?? `Module ${cfg.module_id}`}`;
        case 'photo':
            return `Photo - ${(meta as any).displayed_name ?? `Module ${cfg.module_id}`}`;
        case 'audio':
            return `Audio - ${(meta as any).displayed_name ?? `Module ${cfg.module_id}`}`;
        case 'weather':
            return `Weather - Zip: ${(meta as any).zipcode}`;
        case 'sports':
            return `Sports - ${(meta as any).language || 'Sports News'}`;
        case 'news':
            return 'Good News';
        case 'messages': {
            const msg = (meta as any).message;
            return msg ? (msg.length > 30 ? `Message - ${msg.slice(0, 30)}...` : `Message - ${msg}`) : 'Message';
        }
        case 'event':
            return `Event - ${(meta as any).text}`;
        case 'carepartner_location':
            return `Care Partner Location - ${(meta as any).location}`;
        case 'caregiver_location':
            return `Caregiver Location - ${(meta as any).location}`;
        default:
            return `Module ${cfg.module_id}`;
    }
}
