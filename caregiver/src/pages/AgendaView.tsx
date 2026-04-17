/**
 * AgendaView.tsx
 * 
 * User: Caregiver
 * 
 * Main agenda management screen for a selected date.
 * 
 * Responsibilities:
 * - Fetch and display agenda configurations for a given day
 * - Separate and manage timed vs non-timed items
 * - Allow navigation between days
 * - Support creating and editing agenda configurations
 * - Integrate with Library for module selection
 * - Handle care partner selection context
 * 
 * Data Flow:
 * Calendar → AgendaView → Library (select module) → ConfigurationForm → Save → Refresh
 */

import { useNavigate, useParams } from 'react-router-dom';
import { useState, useEffect, useMemo } from 'react';
import AgendaList, { type ConfigurationWithMetadata } from '../components/AgendaList';
import ConfigurationForm from '../components/ConfigurationForm';
import Library from './Library';
import { useAuth } from '../context/AuthContext';

/* ============================
   Type Definitions
============================ */

/**
 * Represents a selected module coming from the Library
 * Used when creating a new configuration
 */
interface SelectedModule {
    moduleId: number;
    moduleType: string;
    label: string;
}

/* ============================
   Main Component
============================ */
export default function AgendaView() {
    /* ============================
       Routing / Context
    ============================ */

    /**
     * Date comes from route param (format: YYYY-MM-DD)
     */
    const { date } = useParams();

    /**
     * Used for navigation between days
     */
    const navigate = useNavigate();

    /**
     * Auth context:
     * - user (caregiver)
     * - active care partner
     */
    const { user, activeCarePartnerId, setActiveCarePartnerId } = useAuth();

    /* ============================
       State
    ============================ */

    /**
     * All agenda configurations for the selected date
     */
    const [items, setItems] = useState<ConfigurationWithMetadata[]>([]);

    /**
     * Controls Library modal visibility
     */
    const [isLibraryModalOpen, setIsLibraryModalOpen] = useState(false);

    /**
     * Loading & error states for API calls
     */
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // For the create flow: library → config form
    /**
     * Stores selected module when creating a new configuration
     */
    const [selectedModule, setSelectedModule] = useState<SelectedModule | null>(null);

    // For the edit flow: clicking an existing item
    /**
     * Stores the configuration being edited
     */
    const [editingConfig, setEditingConfig] = useState<ConfigurationWithMetadata | null>(null);

    // Careapartners associated with logged in caregiver
    /**
     * List of care partners available to the caregiver
     */
    const [carepartners, setCarepartners] = useState<{ id: number; name: string }[]>([]);

    /* ============================
       Data Fetching
    ============================ */

    /**
     * Fetch agenda configurations for the current date
     */
    const fetchData = async () => {
        try {
            setLoading(true);
            setError(null);

            if (!activeCarePartnerId || !user?.id) {
                setError("Missing caregiver or carepartner");
                setLoading(false);
                return;
            }

            const { dayOfMonth, month, dayOfWeek } = currentDateToComponents();
            /**
             * Guard against invalid date
             */
            if (dayOfMonth === 0) {
                setLoading(false);
                return;
            }
            if (!user?.id || !user?.carepartners) {
                setLoading(false);
                return;
            }

            /**
             * Build query params for backend
             */
            const params = new URLSearchParams({
                caregiver_id: String(user?.id),
                carepartner_id: String(activeCarePartnerId),
                day: String(dayOfMonth),
                month: String(month),
                day_of_week: String(dayOfWeek),
                date: date!,
            });

            const API_URL = import.meta.env.VITE_API_URL;

            /**
             * Fetch configurations for the selected day
             */
            const res = await fetch(`${API_URL}/configurations/day?${params}`);
            if (!res.ok) throw new Error('Failed to fetch configurations');
            const data: ConfigurationWithMetadata[] = await res.json();
            setItems(data || []);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An error occurred';
            setError(errorMessage);
            console.error('Error fetching data:', err);
        } finally {
            setLoading(false);
        }
    };

    /* ============================
        Effects
    ============================ */

    /**
     * Refetch data when:
     * - date changes
     * - care partner changes
     * - user changes
     */
    useEffect(() => {
        fetchData();
    }, [date, activeCarePartnerId, user]);

    /**
     * Fetch all care partners for the current caregiver
     */
    useEffect(() => {
        const fetchCarepartners = async () => {
            if (!user?.id) return;

            try {
                const API_URL = import.meta.env.VITE_API_URL;
                const res = await fetch(`${API_URL}/caregivers/${user.id}/carepartners`);
                if (!res.ok) throw new Error("Failed to fetch carepartners");

                const data = await res.json();
                setCarepartners(data);
            } catch (err) {
                console.error("Failed to fetch carepartners:", err);
            }
        };

        fetchCarepartners();
    }, [user]);

    /* ============================
       Date Utilities
    ============================ */

    /**
     * Converts current date string into components used by backend
     */
    const currentDateToComponents = () => {
        if (!date) return { dayOfMonth: 0, month: 0, dayOfWeek: 0 };
        const currentDate = new Date(date + 'T00:00:00');
        const dayOfWeek = currentDate.getDay();
        const dayOfMonth = currentDate.getDate();
        const month = currentDate.getMonth() + 1;
        return { dayOfMonth, month, dayOfWeek };
    }

    /* ============================
       Navigation Handlers
    ============================ */

    /**
     * Navigate to previous day
     */
    const handlePreviousDay = () => {
        if (!date) return;
        const currentDate = new Date(date + 'T00:00:00');
        currentDate.setDate(currentDate.getDate() - 1);
        const y = currentDate.getFullYear();
        const m = String(currentDate.getMonth() + 1).padStart(2, '0');
        const d = String(currentDate.getDate()).padStart(2, '0');
        const newDate = `${y}-${m}-${d}`;
        navigate(`/agenda/${newDate}`);
    };

    /**
     * Navigate to next day
     */
    const handleNextDay = () => {
        if (!date) return;
        const currentDate = new Date(date + 'T00:00:00');
        currentDate.setDate(currentDate.getDate() + 1);
        const y = currentDate.getFullYear();
        const m = String(currentDate.getMonth() + 1).padStart(2, '0');
        const d = String(currentDate.getDate()).padStart(2, '0');
        const newDate = `${y}-${m}-${d}`;
        navigate(`/agenda/${newDate}`);
    };

    /* ============================
       Interaction Handlers
    ============================ */

    /**
     * Called when a module is selected from the Library
     */
    const handleModuleSelect = (module: SelectedModule) => {
        setIsLibraryModalOpen(false);
        setSelectedModule(module);
    };

    /**
     * Called when an existing item is clicked (edit mode)
     */
    const handleItemClick = (item: ConfigurationWithMetadata) => {
        setEditingConfig(item);
    };

    /**
     * Called after a configuration is saved (create/update)
     */
    const handleConfigSaved = () => {
        setSelectedModule(null);
        setEditingConfig(null);
        fetchData(); // refresh the agenda list
    };

    /* ============================
       Derived Data
    ============================ */

    /**
     * Non-timed agenda items
     */
    const regularItems = useMemo(() => items.filter(i => !i.timed), [items]);

    /**
     * Timed agenda items (sorted by time)
     */
    const timedItems = useMemo(
        () => items.filter(i => i.timed).sort((a, b) => (a.time ?? '').localeCompare(b.time ?? '')),
        [items],
    );

    /**
     * Update only regular items
     */
    const setRegularItems = (newRegular: ConfigurationWithMetadata[]) => {
        setItems(prev => [...prev.filter(i => i.timed), ...newRegular]);
    };

    /**
     * Update only timed items
     */
    const setTimedItems = (newTimed: ConfigurationWithMetadata[]) => {
        setItems(prev => [...prev.filter(i => !i.timed), ...newTimed]);
    };

    /* ============================
       Render
    ============================ */
    return (
        <div className="min-h-screen bg-white">
            {loading && (
                <div className="flex items-center justify-center min-h-screen">
                    <p className="text-xl text-gray-600">Loading agenda...</p>
                </div>
            )}

            {error && (
                <div className="flex items-center justify-center min-h-screen">
                    <div className="bg-red-50 border-2 border-red-400 rounded-lg p-6">
                        <p className="text-red-700 font-semibold">Error loading data</p>
                        <p className="text-red-600 text-sm mt-2">{error}</p>
                    </div>
                </div>
            )}

            {!loading && !error && (
                <div className="px-8 py-8">
                    <div className="flex items-center justify-between mb-8">
                        <button onClick={handlePreviousDay} className="text-4xl text-gray-800 hover:text-gray-600">&lt;</button>
                        <h2 className="text-4xl font-light text-gray-800">{date} Agenda for {activeCarePartnerId ? carepartners.find(cp => cp.id === activeCarePartnerId)?.name : "..."}</h2>
                        <button onClick={handleNextDay} className="text-4xl text-gray-800 hover:text-gray-600">&gt;</button>
                    </div>

                    <div className="flex gap-8 mb-8">
                        {/* <div className="flex items-center gap-4">
                            <label className="text-gray-700 font-medium">View:</label>
                            <button className="px-6 py-2 bg-gray-200 border-b-2 border-gray-400 rounded text-gray-800 flex items-center gap-2 hover:bg-gray-300">
                                Daily
                                <span>∨</span>
                            </button>
                        </div> */}

                        <div className="flex items-center gap-4">
                            <label className="text-gray-700 font-medium">Care Partner:</label>

                            <select
                                value={activeCarePartnerId ?? ""}
                                onChange={(e) => setActiveCarePartnerId(Number(e.target.value))}
                                className="px-6 py-2 bg-gray-200 border-b-2 border-gray-400 rounded text-gray-800 hover:bg-gray-300"
                            >
                                {carepartners.map((cp) => (
                                    <option key={cp.id} value={cp.id}>
                                        {cp.name}
                                    </option>
                                ))}
                            </select>
                        </div>
                    </div>

                    {/* Regular (Non-Timed) Items */}
                    <div className="mb-8">
                        <h3 className="text-2xl font-semibold text-gray-700 mb-4">Agenda Items</h3>
                        <AgendaList items={regularItems} setItems={setRegularItems} onItemClick={handleItemClick} />
                    </div>

                    {/* Scheduled (Timed) Items */}
                    {timedItems.length > 0 && (
                        <div>
                            <h3 className="text-2xl font-semibold text-teal-700 mb-4">Scheduled Items</h3>
                            <AgendaList items={timedItems} setItems={setTimedItems} onItemClick={handleItemClick} />
                        </div>
                    )}

                    <div className="flex items-center justify-between mt-12">
                        <button className="text-3xl text-gray-600 hover:text-gray-800">⚙</button>
                        <div className="flex gap-6">
                            <button className="px-8 py-3 border-2 border-cyan-400 text-cyan-400 rounded-lg font-semibold hover:bg-cyan-50">
                                Preview
                            </button>
                            <button onClick={() => setIsLibraryModalOpen(true)} className="px-8 py-3 border-2 border-cyan-400 bg-white text-cyan-400 rounded-lg text-3xl font-bold hover:bg-cyan-50">
                                +
                            </button>
                        </div>
                    </div>

                    {/* Library Picker Modal */}
                    {isLibraryModalOpen && (
                        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                            <div className="bg-white rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto relative">
                                <button
                                    onClick={() => setIsLibraryModalOpen(false)}
                                    className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 text-2xl z-10"
                                >
                                    ✕
                                </button>
                                <div className="p-8">
                                    <h2 className="text-2xl font-semibold text-center mb-2 text-teal-600">Select a Module to Add</h2>
                                    <p className="text-center text-gray-500 mb-4 text-sm">Click on an item to add it to the agenda</p>
                                    <Library onSelect={handleModuleSelect} />
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Configuration Form Modal — Create Mode */}
                    {selectedModule && (
                        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                            <div className="bg-white rounded-lg shadow-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto relative">
                                <button
                                    onClick={() => setSelectedModule(null)}
                                    className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 text-2xl z-10"
                                >
                                    ✕
                                </button>
                                <div className="p-8">
                                    <ConfigurationForm
                                        selectedModule={selectedModule}
                                        currentItemCount={items.length}
                                        onSave={handleConfigSaved}
                                        onCancel={() => setSelectedModule(null)}
                                        currentDate={date}
                                    />
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Configuration Form Modal — Edit Mode */}
                    {editingConfig && (
                        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                            <div className="bg-white rounded-lg shadow-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto relative">
                                <button
                                    onClick={() => setEditingConfig(null)}
                                    className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 text-2xl z-10"
                                >
                                    ✕
                                </button>
                                <div className="p-8">
                                    <ConfigurationForm
                                        existingConfig={editingConfig}
                                        currentItemCount={items.length}
                                        onSave={handleConfigSaved}
                                        onCancel={() => setEditingConfig(null)}
                                        currentDate={date}
                                    />
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}