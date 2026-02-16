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
}

export default function AgendaList({ items, setItems }: AgendaListProps) {
    const [draggedIndex, setDraggedIndex] = React.useState<number | null>(null);

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
            case 'PhotoMedia':
            case 'video':
            case 'VideoMedia':
            case 'audio':
            case 'AudioMedia':
                return videoCameraIcon;
            case 'weather':
            case 'WeatherWidget':
            case 'news':
            case 'NewsWidget':
            case 'sports':
                return widgetIcon;
            case 'message':
            case 'MessagePrompt':
            case 'event':
            case 'EventPrompt':
            case 'carepartner_location':
            case 'CarepartnerLocationPrompt':
            case 'caregiver_location':
            case 'CaregiverLocationPrompt':
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

        setItems(reordered);
        setDraggedIndex(null);
    };

    const handleDragEnd = () => {
        setDraggedIndex(null);
        console.log(items);
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
                    <button className="flex-1 flex items-center gap-4 bg-slate-100 p-2 rounded-lg">
                        <img src={getModuleIcon(item.module_type)} alt="module icon" className="w-12 h-12 rounded" />
                        <div className="flex-1 flex">
                            <div className="text-gray-800 font-medium text-2xl">{getModuleTitle(item)}</div>
                        </div>
                        <span className="text-gray-700 text-2xl">&gt;</span>
                    </button>
                </div>
            ))}
        </div>
    );
}
