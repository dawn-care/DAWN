import { Swiper, SwiperSlide } from "swiper/react";
import { Navigation, Pagination, EffectCoverflow } from "swiper/modules";

import "swiper/css";
import "swiper/css/navigation";
import "swiper/css/pagination";
import "swiper/css/effect-coverflow";

export default function Home() {
  const images = [
    "/assets/image1.jpg",
    "/assets/image2.jpg",
    "/assets/image3.jpg",
    "/assets/image4.jpg",
    "/assets/image5.jpg",
  ];

  return (
    <div className="w-full flex flex-col items-center">
      <div className="w-full max-w-4xl mt-8 px-4">
        <div className="relative">
          <Swiper
            modules={[Navigation, Pagination, EffectCoverflow]}
            effect="coverflow"
            centeredSlides={true}
            slidesPerView={1.5}
            slidesPerGroup={1}
            navigation
            pagination={{ clickable: true }}
            loop={true}
            observer={true}
            observeParents={true}
            coverflowEffect={{
              rotate: 0,
              stretch: 0,
              depth: 120,
              modifier: 1.2,
              slideShadows: false,
            }}
            className="py-6"
          >
            {images.map((src, idx) => (
              <SwiperSlide key={idx}>
                <div className="flex justify-center">
                  <img
                    src={src}
                    alt={`Slide ${idx + 1}`}
                    className="max-h-96 object-contain rounded-lg"
                  />
                </div>
              </SwiperSlide>
            ))}
          </Swiper>
        </div>
      </div>

      <div className="max-w-3xl mx-auto mt-12 px-6 text-center">
        <h2 className="text-3xl font-semibold text-gray-900 mb-6">
          Help Your Loved Ones Reconnect With Themselves
        </h2>

        <p className="text-base text-gray-700 mb-8 leading-relaxed">
          Our mission is to support individuals living with dementia by helping
          them reconnect with their family, their memories, and the moments that
          shaped their lives. Through meaningful reminders and shared
          experiences, we aim to bring comfort, familiarity, and love back into
          everyday life.
        </p>
      </div>

    </div>
  );
}