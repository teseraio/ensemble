/* This example requires Tailwind CSS v2.0+ */
import { AnnotationIcon, GlobeAltIcon, LightningBoltIcon, ScaleIcon } from '@heroicons/react/outline'

export default function Example({title, abstract, features}) {
  return (
    <div className="py-12 pb-20 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="lg:text-center">
          <p className="mt-2 text-3xl leading-8 font-extrabold tracking-tight text-gray-900 sm:text-4xl">
            {title}
          </p>
        </div>

        <div className="mt-10">
          <dl className="space-y-10 md:space-y-0 md:grid md:grid-cols-2 md:gap-x-8 md:gap-y-10">
            {features.map((feature) => (
              <div key={feature.name} className="relative bg-white2 p-8">
                <dt>
                  <div className="md:absolute p-2 flex text-white">
                    <feature.icon className="h-16 w-16" aria-hidden="true" />
                  </div>
                  <p className="md:ml-24 text-lg leading-6 font-bold text-black">{feature.name}</p>
                </dt>
                <dd className="mt-2 md:ml-24 text-lg text-tgrey-white">{feature.description}</dd>
              </div>
            ))}
          </dl>
        </div>
      </div>
    </div>
  )
}
