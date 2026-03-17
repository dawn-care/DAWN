import React from 'react';
import videoCameraIcon from '../assets/module_symbols/video-camera-svgrepo-com.svg';
import widgetIcon from '../assets/module_symbols/widget-svgrepo-com.svg';
import promptIcon from '../assets/module_symbols/prompt-edit-svgrepo-com.svg';

// Metadata shapes returned inline by the backend
export type PhotoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type VideoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type AudioMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type WeatherMetadata = { id: number; zipcode: string };
export type NewsMetadata = { id: number; category: string; language: string; country: string };
export type MessageMetadata = { id: number; message: string; image_uri: string };
export type EventMetadata = { id: number; list_id: number; time: string; text: string };
export type CarepartnerLocationMetadata = { id: number; location: string; message: string; image_uri: string };
export type CaregiverLocationMetadata = { id: number; location: string; message: string; image_uri: string };

// Matches the ConfigurationWithMetadata Go struct returned by GetConfigurationsByDayAndCarepartner
export type ConfigurationWithMetadata = {
    id: number;
    caregiver_id: number;
    carepartner_id: number;
    module_id: number;
    occurence_type: string;
    occurence_definition: string[];
    position: number;
    duration: number;
    module_type: string;
    metadata_id: number;
    metadata: PhotoMetadata | VideoMetadata | AudioMetadata | WeatherMetadata
        | NewsMetadata | MessageMetadata | EventMetadata
        | CarepartnerLocationMetadata | CaregiverLocationMetadata | null;
};

interface AgendaListProps {
    items: ConfigurationWithMetadata[];
    setItems: (items: ConfigurationWithMetadata[]) => void;
    onItemClick?: (item: ConfigurationWithMetadata) => void;
}

export default function AgendaList({ items, setItems, onItemClick }: AgendaListProps) {
    const [draggedIndex, setDraggedIndex] = React.useState<number | null>(null);
    const [pendingReorder, setPendingReorder] = React.useState<{
        reordered: ConfigurationWithMetadata[];
        movedTitle: string;
        fromPos: number;
        toPos: number;
    } | null>(null);
    // Keep a snapshot so we can revert on cancel
    const [preReorderItems, setPreReorderItems] = React.useState<ConfigurationWithMetadata[] | null>(null);

    const getModuleTitle = (item: ConfigurationWithMetadata): string => {
        const meta = item.metadata;
        if (!meta) return `Module ${item.module_id}`;

        switch (item.module_type) {
            case 'video':
                return "Video  - " + ((meta as VideoMetadata).displayed_name ?? `Module ${item.module_id}`);
            case 'photo':
                return "Photo  - " + ((meta as PhotoMetadata).displayed_name ?? `Module ${item.module_id}`);
            case 'audio':
                return "Audio  - " + ((meta as AudioMetadata).displayed_name ?? `Module ${item.module_id}`);
            case 'weather':
                return `Weather - Zip Code: ${(meta as WeatherMetadata).zipcode}`;
            case 'news':
            case 'sports':
                return `News - Category: ${(meta as NewsMetadata).category}`;
            case 'message':
                return 'Message';
            case 'event':
                return `Event - ${(meta as EventMetadata).text}`;
            case 'carepartner_location':
                return `Location - ${(meta as CarepartnerLocationMetadata).location}`;
            case 'caregiver_location':
                return `Location - ${(meta as CaregiverLocationMetadata).location}`;
            default:
                return `Module ${item.module_id}`;
        }
    };

    const getModuleIcon = (moduleType: string): string => {
        // Map detailed module_type values to icon categories
        switch (moduleType) {
            case 'photo':
            case 'video':
            case 'audio':
                return videoCameraIcon;
            case 'weather':
            case 'news':
            case 'sports':
                return widgetIcon;
            case 'message':
            case 'event':
            case 'carepartner_location':
            case 'caregiver_location':
                return promptIcon;
            default:
                return widgetIcon;
        }
    };

    /* Handlers for drag and drop */
    const handleDragStart = (index: number) => {
        setDraggedIndex(index);
    };

    const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
    };

    const handleDrop = (dropIndex: number) => {
        if (draggedIndex === null || draggedIndex === dropIndex) return;

        const targetItem = items[dropIndex];
        const draggedItem = items[draggedIndex];

        const updatedItems = items.map((item, index) => {
            // Dragging UP
            if (dropIndex < draggedIndex) {
                if (index === draggedIndex) {
                    return {
                        ...item,
                        position: targetItem.position
                    };
                }

                if (index >= dropIndex && index < draggedIndex) {
                    return {
                        ...item,
                        position: item.position + 1
                    };
                }
            }

            // Dragging DOWN
            if (dropIndex > draggedIndex) {
                if (index === draggedIndex) {
                    return {
                        ...item,
                        position: targetItem.position
                    };
                }

                if (index > draggedIndex && index <= dropIndex) {
                    return {
                        ...item,
                        position: item.position - 1
                    };
                }
            }

            return item;
        });

        // Reorder array visually
        const reordered = [...updatedItems];
        const [removed] = reordered.splice(draggedIndex, 1);
        reordered.splice(dropIndex, 0, removed);

        // Save snapshot for potential revert, apply reorder optimistically, and show confirmation
        setPreReorderItems([...items]);
        setItems(reordered);
        setPendingReorder({
            reordered,
            movedTitle: getModuleTitle(draggedItem),
            fromPos: draggedItem.position,
            toPos: targetItem.position,
        });
        setDraggedIndex(null);
    };

    const handleDragEnd = () => {
        setDraggedIndex(null);
    };

    const confirmReorder = async () => {
        if (!pendingReorder) return;
        const reordered = pendingReorder.reordered;

        try {
            const configEntries = reordered.map((item) => ({
                id: item.id,
                module_id: item.module_id,
                occurence_type: item.occurence_type,
                occurence_definition: item.occurence_definition ?? [],
                position: item.position,
                duration: item.duration,
            }));

            const API_URL = import.meta.env.VITE_API_URL + "/api";
            const res = await fetch(`${API_URL}/configurations/insert`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    carepartner_id: 1,
                    json_config: JSON.stringify(configEntries),
                }),
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed to update positions');
            }
        } catch (err) {
            console.error('Reorder failed, reverting:', err);
            if (preReorderItems) {
                setItems(preReorderItems);
            }
        } finally {
            setPendingReorder(null);
            setPreReorderItems(null);
        }
    };

    const cancelReorder = () => {
        if (preReorderItems) {
            setItems(preReorderItems);
        }
        setPendingReorder(null);
        setPreReorderItems(null);
    };

    return (
        <div className="space-y-4">
            {items.map((item, index) => (
                <div 
                    key={item.id || index}
                    onDragOver={handleDragOver}
                    onDrop={() => handleDrop(index)}
                    className={`flex items-center gap-1 bg-gray-300 px-2 py-2 rounded-lg transition-opacity ${
                        draggedIndex === index ? 'opacity-50' : 'opacity-100'
                    }`}
                >
                    <div
                        draggable
                        onDragStart={() => handleDragStart(index)}
                        onDragEnd={handleDragEnd}
                        className="cursor-move flex items-center gap-1 px-2"
                    >
                        {/* Six dot drag handle */}
                        <div className="flex flex-col gap-1">
                            <div className="flex gap-1">
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                            </div>
                            <div className="flex gap-1">
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                            </div>
                            <div className="flex gap-1">
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                            </div>
                        </div>
                    </div>
                    <button
                        className="flex-1 flex items-center gap-4 bg-slate-100 p-2 rounded-lg hover:bg-slate-200 transition"
                        onClick={() => onItemClick?.(item)}
                    >
                        <img src={getModuleIcon(item.module_type)} alt="module icon" className="w-12 h-12 rounded" />
                        <div className="flex-1 flex">
                            <div className="text-gray-800 font-medium text-2xl">{getModuleTitle(item)}</div>
                        </div>
                        <span className="text-gray-700 text-2xl">&gt;</span>
                    </button>
                </div>
            ))}

            {/* Reorder Confirmation Popup */}
            {pendingReorder && (
                <div className="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50">
                    <div className="bg-white rounded-xl shadow-xl p-6 max-w-md w-full mx-4">
                        <h3 className="text-xl font-semibold text-gray-800 mb-3">Confirm Position Change</h3>
                        <p className="text-gray-600 mb-6">
                            Move <span className="font-medium text-gray-800">"{pendingReorder.movedTitle}"</span>?
                        </p>
                        <div className="flex gap-3 justify-end">
                            <button
                                onClick={cancelReorder}
                                className="px-5 py-2 border border-gray-300 text-gray-600 rounded-lg font-medium hover:bg-gray-50"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={confirmReorder}
                                className="px-5 py-2 bg-teal-500 text-white rounded-lg font-medium hover:bg-teal-600"
                            >
                                Confirm
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
